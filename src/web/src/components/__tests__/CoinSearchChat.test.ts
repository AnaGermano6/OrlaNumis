import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import CoinSearchChat from '../CoinSearchChat.vue'

const mockShowAlert = vi.fn()
const mockNewChat = vi.fn().mockResolvedValue(true)

vi.mock('@/api/client', () => ({
  getApiErrorMessage: () => '',
}))

vi.mock('@/composables/useDialog', () => ({
  useDialog: () => ({ showAlert: mockShowAlert }),
}))

vi.mock('@/composables/useCoinSearchChat', () => ({
  useCoinSearchChat: () => ({
    messages: ref([
      { role: 'user', content: 'Which coins are missing an era?' },
      { role: 'assistant', content: '# Missing Era Report\n\n- Coin #1 needs an era.' },
    ]),
    input: ref(''),
    loading: ref(false),
    addingIdx: ref(null),
    addedSet: ref(new Set()),
    savedShows: ref(new Set()),
    savingShow: ref(null),
    conversationId: ref(null),
    saving: ref(false),
    saveLabel: ref('Save'),
    newChat: mockNewChat,
    newChatDisabled: ref(false),
    providerConfigured: ref(true),
    categoryEraConfirmRequest: ref(null),
    copilotActive: ref(false),
    copilotRun: ref(null),
    copilotPlan: ref([]),
    copilotTools: ref([]),
    copilotClarification: ref(null),
    copilotCanCancel: ref(false),
    copilotCanResume: ref(false),
    copilotTruncated: ref(false),
    copilotError: ref(''),
    copilotCancelling: ref(false),
    copilotResuming: ref(false),
    chooseCategoryEraConfirmation: vi.fn(),
    cancelCategoryEraConfirmation: vi.fn(),
    sendMessage: vi.fn(),
    cancelCopilotRun: vi.fn(),
    resumeCopilotRun: vi.fn(),
    sendExample: vi.fn(),
    sendPortfolioAnalysis: vi.fn(),
    handleSave: vi.fn(),
    addToWishlist: vi.fn(),
    confirmCollectionProposal: vi.fn(),
    cancelCollectionProposalMessage: vi.fn(),
    pickDisambiguationCandidate: vi.fn(),
    formatMessage: (message: string) => message,
    isCoinShowResults: () => false,
    saveShowToCalendar: vi.fn(),
  }),
}))

describe('CoinSearchChat', () => {
  it('starts a new chat from the header', async () => {
    const wrapper = mount(CoinSearchChat)
    await wrapper.findAll('button').find(button => button.text() === 'New Chat')!.trigger('click')
    await flushPromises()
    expect(mockNewChat).toHaveBeenCalledOnce()
    wrapper.unmount()
  })

  it('offers no saving of answers to Notes, which are gone', () => {
    const wrapper = mount(CoinSearchChat, {
      global: { stubs: { CoinSuggestionGrid: true, CoinShowResultsGrid: true, CategoryEraConfirmModal: true } },
    })

    expect(wrapper.text()).toContain('Coin Agent')
    expect(wrapper.text()).toContain('Missing Era Report')
    expect(wrapper.findAll('button').some((button) => button.text() === 'Save to Notes')).toBe(false)
    wrapper.unmount()
  })
})
