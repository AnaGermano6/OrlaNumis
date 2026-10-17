"""Provider fan-out tests (T074) — contracts/agent-internal-contract.md §6.

Verifies bounded concurrency (`max_concurrency`), per-provider timeout ->
`timed_out`, and that one failing provider does not prevent the others
(or the overall pipeline) from reaching a result (partial success).
"""

import asyncio

import pytest

from app.models.requests import DeepIdentifyBounds, DeepProviderCatalogEntry
from app.models.responses import ProviderEvidence
from app.teams.deep_identification.graph import provider_fanout_node


def _bounds(max_concurrency=2, provider_timeout_s=1, max_providers=5) -> DeepIdentifyBounds:
    return DeepIdentifyBounds(
        max_providers=max_providers,
        max_concurrency=max_concurrency,
        provider_timeout_s=provider_timeout_s,
        total_timeout_s=60,
        recursion_limit=10,
    )


@pytest.mark.asyncio
async def test_fanout_includes_non_automatable_providers_even_if_not_selected(monkeypatch):
    catalog = [
        DeepProviderCatalogEntry(provider="ngc", automatable=False, link_out="https://www.ngccoin.com/verify/"),
        DeepProviderCatalogEntry(provider="rpc", automatable=False, reason="no_public_api"),
    ]
    state = {
        "catalog": catalog,
        "bounds": _bounds(),
        "quick_evidence": None,
        "notes": "",
        "selected": [],  # router selected nothing automatable
        "skipped": [],
    }

    result = await provider_fanout_node(state, tools=None)

    providers_seen = {row.provider for row in result["evidence"]}
    assert providers_seen == {"ngc", "rpc"}


@pytest.mark.asyncio
async def test_fanout_respects_max_concurrency(monkeypatch):
    concurrent = 0
    max_seen = 0
    lock = asyncio.Lock()

    async def fake_run(entry, tools, quick_evidence, notes, hypothesis=None):
        nonlocal concurrent, max_seen
        async with lock:
            concurrent += 1
            max_seen = max(max_seen, concurrent)
        await asyncio.sleep(0.05)
        async with lock:
            concurrent -= 1
        return ProviderEvidence(provider=entry.provider, status="no_match", automatable=True, call_count=1)

    import app.teams.deep_identification.graph as graph_module

    # Only Numista is automated today: rpc stands in for a second automated provider.
    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "numista", fake_run)
    monkeypatch.delitem(graph_module._TRIVIAL_PROVIDER_NODES, "rpc")
    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "rpc", fake_run)

    catalog = [
        DeepProviderCatalogEntry(provider=name, automatable=True) for name in ("numista", "rpc")
    ]
    state = {
        "catalog": catalog,
        "bounds": _bounds(max_concurrency=1),
        "quick_evidence": None,
        "notes": "",
        "selected": ["numista", "rpc"],
        "skipped": [],
    }

    await provider_fanout_node(state, tools=object())

    assert max_seen == 1


@pytest.mark.asyncio
async def test_fanout_provider_timeout_yields_timed_out(monkeypatch):
    async def slow_run(entry, tools, quick_evidence, notes, hypothesis=None):
        await asyncio.sleep(10)
        return ProviderEvidence(provider=entry.provider, status="no_match", automatable=True)

    import app.teams.deep_identification.graph as graph_module

    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "numista", slow_run)

    catalog = [DeepProviderCatalogEntry(provider="numista", automatable=True)]
    state = {
        "catalog": catalog,
        "bounds": _bounds(provider_timeout_s=1),
        "quick_evidence": None,
        "notes": "",
        "selected": ["numista"],
        "skipped": [],
    }

    result = await provider_fanout_node(state, tools=object())

    assert result["evidence"][0].status == "timed_out"
    assert result["evidence"][0].error_kind == "timeout"


@pytest.mark.asyncio
async def test_fanout_one_failing_provider_does_not_block_others(monkeypatch):
    async def failing_run(entry, tools, quick_evidence, notes, hypothesis=None):
        raise RuntimeError("boom")

    async def ok_run(entry, tools, quick_evidence, notes, hypothesis=None):
        return ProviderEvidence(provider=entry.provider, status="contributed", automatable=True, call_count=1)

    import app.teams.deep_identification.graph as graph_module

    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "numista", failing_run)
    monkeypatch.delitem(graph_module._TRIVIAL_PROVIDER_NODES, "rpc")
    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "rpc", ok_run)

    catalog = [
        DeepProviderCatalogEntry(provider="numista", automatable=True),
        DeepProviderCatalogEntry(provider="rpc", automatable=True),
    ]
    state = {
        "catalog": catalog,
        "bounds": _bounds(),
        "quick_evidence": None,
        "notes": "",
        "selected": ["numista", "rpc"],
        "skipped": [],
    }

    result = await provider_fanout_node(state, tools=object())

    by_provider = {row.provider: row for row in result["evidence"]}
    assert by_provider["numista"].status == "failed"
    assert by_provider["rpc"].status == "contributed"


@pytest.mark.asyncio
async def test_fanout_rpc_trivial_path_unaffected_by_dict_split():
    """RPC is the sole trivial (single-arg, synchronous) node; it emits its
    typed unavailable row with zero calls."""
    import app.teams.deep_identification.graph as graph_module

    assert set(graph_module._TRIVIAL_PROVIDER_NODES) == {"rpc"}

    catalog = [DeepProviderCatalogEntry(provider="rpc", automatable=False, reason="no_public_api")]
    state = {
        "catalog": catalog,
        "bounds": _bounds(),
        "quick_evidence": None,
        "notes": "",
        "selected": [],
        "skipped": [],
    }

    result = await provider_fanout_node(state, tools=None)

    row = result["evidence"][0]
    assert row.provider == "rpc"
    assert row.status == "unavailable"
    assert row.call_count == 0


@pytest.mark.asyncio
async def test_fanout_hung_provider_still_yields_already_completed_evidence(monkeypatch):
    """T036: a provider hung past its timeout must not strand the evidence
    other providers already produced — the timed-out row is typed and the
    completed provider's contribution survives (incremental on_result)."""
    async def hung_provider(entry, tools, quick_evidence, notes, hypothesis=None):
        await asyncio.sleep(10)
        return ProviderEvidence(provider="rpc", status="contributed", automatable=True)

    async def fast_numista(entry, tools, quick_evidence, notes, hypothesis=None):
        return ProviderEvidence(provider="numista", status="contributed", automatable=True, call_count=1)

    import app.teams.deep_identification.graph as graph_module

    monkeypatch.delitem(graph_module._TRIVIAL_PROVIDER_NODES, "rpc")
    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "rpc", hung_provider)
    monkeypatch.setitem(graph_module._AUTOMATED_PROVIDER_NODES, "numista", fast_numista)

    accumulated: list = []
    catalog = [
        DeepProviderCatalogEntry(provider="rpc", automatable=True),
        DeepProviderCatalogEntry(provider="numista", automatable=True),
    ]
    state = {
        "catalog": catalog,
        "bounds": _bounds(provider_timeout_s=1, max_concurrency=2),
        "quick_evidence": None,
        "notes": "",
        "selected": ["rpc", "numista"],
        "skipped": [],
    }

    result = await provider_fanout_node(state, tools=object(), on_result=lambda row: accumulated.append(row))

    by_provider = {row.provider: row.status for row in result["evidence"]}
    assert by_provider["numista"] == "contributed"
    assert by_provider["rpc"] == "timed_out"
    # The completed provider's evidence was accumulated incrementally, before
    # the hung task resolved.
    assert any(r.provider == "numista" and r.status == "contributed" for r in accumulated)


