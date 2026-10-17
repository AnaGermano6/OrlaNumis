import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SetCreationWizard from '@/components/sets/SetCreationWizard.vue'
import { emptyFilters } from '@/utils/setFilters'

const mocks = vi.hoisted(() => ({ getJointIssues: vi.fn() }))
vi.mock('@/api/client', () => mocks)

const JOINT_ISSUES = [
  { year: 2022, title: '35 years of the Erasmus programme', pageUrl: 'https://ecb/2022', coins: [{ country: 'AT', imageUrl: 'https://ecb/at22.jpg' }, { country: 'PT', imageUrl: 'https://ecb/pt22.jpg' }] },
  { year: 2007, title: '50th anniversary of the Treaty of Rome', pageUrl: 'https://ecb/2007', coins: [{ country: 'BE', imageUrl: 'https://ecb/be07.jpg' }, { country: 'DE', imageUrl: 'https://ecb/de07.jpg' }, { country: 'PT', imageUrl: 'https://ecb/pt07.jpg' }] },
]

const stubs = {
  RouterLink: { props: ['to'], template: '<a :href="to" @click="$emit(\'click\')"><slot /></a>' },
  SetSimpleFilters: { name: 'SetSimpleFilters', props: ['modelValue'], emits: ['update:modelValue'], template: '<div data-testid="set-filters" />' },
}

function mountWizard(props = {}) {
  return mount(SetCreationWizard, { attachTo: document.body, props, global: { stubs } })
}

function submitted(wrapper: ReturnType<typeof mountWizard>): Record<string, unknown> {
  const events = wrapper.emitted('submit') as unknown[][]
  expect(events).toBeTruthy()
  return events[0]![0] as Record<string, unknown>
}

const COLLECTION_RULES = [
  { field: 'isWishlist', op: 'eq', value: false },
]

describe('SetCreationWizard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.getJointIssues.mockResolvedValue({ data: { issues: JOINT_ISSUES } })
  })

  it('starts with the ways to make a set, and goes back to them', async () => {
    const wrapper = mountWizard()
    await flushPromises()

    expect(wrapper.findAll('[data-testid^="set-path-"]').map((node) => node.attributes('data-testid'))).toEqual([
      'set-path-catalogue', 'set-path-challenge', 'set-path-filters', 'set-path-manual', 'set-path-goal', 'set-path-agentic',
    ])
    expect(wrapper.find('form').exists()).toBe(false)

    await wrapper.get('[data-testid="set-path-manual"]').trigger('click')
    expect(wrapper.get('[data-testid="set-path-title"]').text()).toBe('Choose the coins')
    await wrapper.get('[data-testid="set-path-back"]').trigger('click')
    expect(wrapper.find('[data-testid="set-paths"]').exists()).toBe(true)
  })

  it('sends the catalogue path to the catalogue, where the set is made', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-catalogue"]').trigger('click')

    const link = wrapper.get('[data-testid="set-open-catalogue"]')
    expect(link.attributes('href')).toBe('/sets/commemoratives')
    expect(wrapper.find('#setName').exists()).toBe(false)
    await link.trigger('click')
    expect(wrapper.emitted('cancel')).toBeTruthy()
  })

  it('makes a challenge a checklist set named after it', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-challenge"]').trigger('click')

    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe('A coin of each euro country')
    await wrapper.get('form').trigger('submit.prevent')

    expect(submitted(wrapper)).toEqual(expect.objectContaining({
      name: 'A coin of each euro country',
      setType: 'checklist',
      checklist: { source: 'challenge', challenge: 'euro-countries' },
    }))
  })

  it('offers the joint issues, each the same coin in many countries', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-challenge"]').trigger('click')
    await wrapper.get('[data-testid="challenge-jointIssue"] input').setValue(true)
    await flushPromises()

    expect(mocks.getJointIssues).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('#jointIssue option').map((option) => option.text())).toEqual([
      '2022 — 35 years of the Erasmus programme (2 countries)',
      '2007 — 50th anniversary of the Treaty of Rome (3 countries)',
    ])
    // The newest at first, shown by its Portuguese version.
    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe('2022 joint issue: 35 years of the Erasmus programme')
    expect(wrapper.get('[data-testid="joint-issue-picture"]').attributes('src')).toBe('https://ecb/pt22.jpg')
    await wrapper.get('#jointIssue').setValue('2007')
    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe('2007 joint issue: 50th anniversary of the Treaty of Rome')
    await wrapper.get('form').trigger('submit.prevent')

    expect(submitted(wrapper)).toEqual(expect.objectContaining({
      setType: 'checklist',
      checklist: { source: 'challenge', challenge: 'joint-issue', year: 2007 },
    }))
  })

  it('says so when the joint issues cannot be loaded, and creates none', async () => {
    mocks.getJointIssues.mockRejectedValue(new Error('offline'))
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-challenge"]').trigger('click')
    await wrapper.get('[data-testid="challenge-jointIssue"] input').setValue(true)
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toBe('The joint issues could not be loaded.')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it("asks the year's series for a country and a year", async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-challenge"]').trigger('click')
    await wrapper.get('[data-testid="challenge-annualSeries"] input').setValue(true)

    const submit = wrapper.get('button[type="submit"]')
    expect(submit.attributes('disabled')).toBeDefined()
    expect(wrapper.findAll('#challengeCountry option').map((option) => option.text())).toContain('Portugal')
    await wrapper.get('#challengeCountry').setValue('PT')
    await wrapper.get('#challengeYear').setValue('1998')
    // Before the euro.
    expect(submit.attributes('disabled')).toBeDefined()
    await wrapper.get('#challengeYear').setValue('2002')
    expect(submit.attributes('disabled')).toBeUndefined()
    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe("Portugal 2002: year's series")
    await wrapper.get('form').trigger('submit.prevent')

    expect(submitted(wrapper)).toEqual(expect.objectContaining({
      setType: 'checklist',
      checklist: { source: 'challenge', challenge: 'annual-series', country: 'PT', year: 2002 },
    }))
  })

  it('makes the year of birth a smart set of the coins of that year', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-challenge"]').trigger('click')
    await wrapper.get('[data-testid="challenge-birthYear"] input').setValue(true)
    await wrapper.get('#birthYear').setValue('1990')
    await wrapper.get('form').trigger('submit.prevent')

    expect(submitted(wrapper)).toEqual(expect.objectContaining({
      name: 'Coins of 1990',
      setType: 'smart',
      smartCriteria: { operator: 'and', rules: [{ field: 'year', op: 'eq', value: 1990 }, ...COLLECTION_RULES] },
    }))
  })

  it('makes filters a smart set, named after them until a name is typed', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-filters"]').trigger('click')
    const filters = wrapper.findComponent({ name: 'SetSimpleFilters' })
    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe('All my coins')

    filters.vm.$emit('update:modelValue', { ...emptyFilters(), country: 'Portugal', denomination: '2 euro', classification: 'commemorative' })
    await flushPromises()
    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe('Portugal · 2 euro · Commemorative')

    await wrapper.get('#setName').setValue('My Portuguese €2')
    filters.vm.$emit('update:modelValue', { ...emptyFilters(), country: 'Portugal', denomination: '2 euro', yearFrom: 2010 })
    await flushPromises()
    expect(wrapper.get<HTMLInputElement>('#setName').element.value).toBe('My Portuguese €2')
    await wrapper.get('form').trigger('submit.prevent')

    expect(submitted(wrapper)).toEqual(expect.objectContaining({
      name: 'My Portuguese €2',
      setType: 'smart',
      smartCriteria: { operator: 'and', rules: [
        { field: 'country', op: 'eq', value: 'Portugal' },
        { field: 'denomination', op: 'eq', value: '2 euro' },
        { field: 'year', op: 'gte', value: 2010 },
        ...COLLECTION_RULES,
      ] },
    }))
  })

  it('makes a set of coins chosen by hand', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-manual"]').trigger('click')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.get('#setName').setValue('  Favourites ')
    await wrapper.get('form').trigger('submit.prevent')

    expect(submitted(wrapper)).toEqual(expect.objectContaining({ name: 'Favourites', setType: 'standard' }))
    expect(submitted(wrapper)).not.toHaveProperty('smartCriteria')
  })

  it('emits an agentic submit payload with the prompt instead of creating a set locally', async () => {
    const wrapper = mountWizard()
    await flushPromises()

    await wrapper.get('[data-testid="set-path-agentic"]').trigger('click')
    await wrapper.find('#agenticPrompt').setValue('  All US Silver Quarters from 1940s to 1960s  ')
    await wrapper.find('#setName').setValue('US Silver Quarters')

    expect(wrapper.text()).toContain('How agentic sets work')
    expect(wrapper.text()).toContain('up to 300 roster slots')
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.find('button[type="submit"]').text()).toBe('Create')

    await wrapper.find('form').trigger('submit.prevent')

    const events = wrapper.emitted('submit') as unknown[][]
    const payload = submitted(wrapper)
    expect(payload.setType).toBe('agentic')
    expect(payload.agenticPrompt).toBe('All US Silver Quarters from 1940s to 1960s')
    expect(events[0]![1]).toBeUndefined()
  })

  it('shows goal completion guidance, and opens a goal set to edit on its form', async () => {
    const wrapper = mountWizard()
    await wrapper.get('[data-testid="set-path-goal"]').trigger('click')

    expect(wrapper.find('#templateId').exists()).toBe(false)
    expect(wrapper.find('#targetCompletionDate').exists()).toBe(false)
    expect(wrapper.text()).toContain('Goal sets track both collection and wishlist members.')
    expect(wrapper.text()).toContain('2 / (2 + 5) = 28.6%')

    const goal = mountWizard({ initialValue: { setType: 'goal', templateId: 'roman-template', targetCompletionDate: '2027-01-01' } })
    await goal.find('#setName').setValue('Roman Goal Set')
    await goal.find('form').trigger('submit.prevent')
    const payload = submitted(goal)
    expect(payload.setType).toBe('goal')
    expect(payload.templateId).toBeUndefined()
    expect(payload.targetCompletionDate).toBeUndefined()
  })
})
