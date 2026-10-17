"""Router node (contracts/agent-internal-contract.md §6, T044-T046).

A **pure, deterministic** function of `(catalog, provider_override, bounds)`
selects which *automatable* providers to query
this run, bounded by `bounds.max_providers`. No LLM call is made — this was
previously an LLM-router step (344 FR-022); it is now a plain evidence-driven
selector (spec FR-013/FR-014, ADR 0012). Non-automatable catalog entries
(`ngc`, `rpc`) are never subject to router
reasoning — they always run (trivially, with no network call) and are
represented directly in the final evidence list by their own provider nodes,
so they are excluded from `selected`/`skipped` here (those two lists describe
only automatable candidates the router did or didn't pick).

`provider_override`, when non-empty, replaces selection reasoning entirely:
only catalog-listed automatable providers named in the override are selected,
and no other automatable provider runs. The override can never introduce a
provider absent from the catalog — Go controls the closed candidate list.

Selection is **inclusion by default** (spec FR-015, RD-7): every automatable,
in-bounds provider is selected.
"""

from app.models.requests import DeepIdentifyBounds, DeepProviderCatalogEntry
from app.teams.deep_identification.merge import PROVIDER_RANK
from app.teams.deep_identification.state import RouterSkip


class RouterDecision:
    """Result of the router step."""

    def __init__(self, selected: list[str], skipped: list[RouterSkip], rationale: str) -> None:
        self.selected = selected
        self.skipped = skipped
        self.rationale = rationale


def _automatable_names(catalog: list[DeepProviderCatalogEntry]) -> list[str]:
    return [entry.provider for entry in catalog if entry.automatable]


def _rank(name: str) -> int:
    try:
        return PROVIDER_RANK.index(name)
    except ValueError:
        return len(PROVIDER_RANK)


def route(
    catalog: list[DeepProviderCatalogEntry],
    provider_override: list[str],
    bounds: DeepIdentifyBounds | None,
) -> RouterDecision:
    """Select which automatable providers to query this run. Pure and
    deterministic (spec FR-014, SC-006): identical inputs always produce an
    identical `selected` set, `skipped` list (with reasons), and
    `rationale`.
    """
    automatable = _automatable_names(catalog)
    if not automatable:
        return RouterDecision(selected=[], skipped=[], rationale="no automatable providers in catalog")

    budget = max(0, bounds.max_providers if bounds else 0)

    if provider_override:
        # Override wins outright — no reasoning, no free-form provider names.
        allowed_override = [p for p in provider_override if p in automatable]
        selected = sorted(allowed_override, key=_rank)[:budget] if budget else []
        skipped = [
            RouterSkip(provider=p, reason="not requested by provider_override")
            for p in automatable
            if p not in selected
        ]
        return RouterDecision(
            selected=selected, skipped=skipped, rationale="provider_override supplied by caller"
        )

    # Inclusion by default (FR-015, RD-7): every automatable provider is a
    # candidate, up to the budget.
    ordered = sorted(automatable, key=_rank)
    selected = ordered[:budget] if budget else []
    selected_set = set(selected)

    skipped = [
        RouterSkip(provider=p, reason="max_providers limit reached") for p in automatable if p not in selected_set
    ]

    rationale = f"inclusion by default: selected {len(selected)} of {len(automatable)} automatable providers"

    return RouterDecision(selected=selected, skipped=skipped, rationale=rationale)
