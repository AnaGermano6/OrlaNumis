// The critical workflows the AI browser exploration may be scoped to. Each one
// names the routes it may visit and the checkpoints its report records.
export interface F013Workflow {
  readonly id: string
  readonly goal: string
  readonly routes: readonly string[]
  readonly checkpoints: readonly string[]
  readonly viewport: Readonly<{ width: number; height: number }>
}

const desktop = Object.freeze({ width: 1440, height: 900 })
const mobile = Object.freeze({ width: 390, height: 844 })

// It describes the existing workflows; it does not copy fixture data.
export const F013_WORKFLOW_INVENTORY = Object.freeze({
  'login-session': Object.freeze({
    id: 'login-session', goal: 'Sign in through the normal login flow and verify the session.',
    routes: Object.freeze(['/login', '/']), checkpoints: Object.freeze(['login-form', 'authenticated-home']), viewport: desktop,
  }),
  'add-coin': Object.freeze({
    id: 'add-coin', goal: 'Add one synthetic coin and verify it appears in the collection.',
    routes: Object.freeze(['/', '/add', '/coin/1']), checkpoints: Object.freeze(['new-coin-form', 'created-coin']), viewport: desktop,
  }),
  'edit-one-field': Object.freeze({
    id: 'edit-one-field', goal: 'Edit one field on a golden fixture coin and verify the saved value.',
    routes: Object.freeze(['/', '/coin/1', '/edit/1']), checkpoints: Object.freeze(['edit-form', 'saved-value']), viewport: desktop,
  }),
  'sets-edit': Object.freeze({
    id: 'sets-edit', goal: 'Add and remove canonical sets on a coin.',
    routes: Object.freeze(['/coin/1']), checkpoints: Object.freeze(['sets-updated']), viewport: desktop,
  }),
  'image-upload-delete': Object.freeze({
    id: 'image-upload-delete', goal: 'Upload and delete a packaged synthetic image.',
    routes: Object.freeze(['/coin/1', '/edit/1']), checkpoints: Object.freeze(['image-uploaded', 'image-deleted']), viewport: desktop,
  }),
  'collection-search-filter': Object.freeze({
    id: 'collection-search-filter', goal: 'Search and filter the canonical golden collection.',
    routes: Object.freeze(['/']), checkpoints: Object.freeze(['search-result', 'filtered-result']), viewport: desktop,
  }),
  'mobile-edit': Object.freeze({
    id: 'mobile-edit', goal: 'Edit a golden fixture coin at a mobile viewport.',
    routes: Object.freeze(['/', '/coin/1', '/edit/1']), checkpoints: Object.freeze(['mobile-form', 'mobile-saved']), viewport: mobile,
  }),
}) as Readonly<Record<string, F013Workflow>>
