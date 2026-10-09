export default {
  agentUnavailable: 'Agent service unavailable. Check the internal agent service configuration.',
  agentCredentialMissing: 'Internal agent service credential is not configured. Check the internal agent service configuration.',
  agentCheckConfig: 'Check the agent service configuration and retry.',
  setBuilder: {
    slotLimit: 'The set-builder request exceeded the workflow slot limit. Try a narrower prompt by date range, mint, ruler, or series, then submit again.',
    rejected: 'The agent rejected the set-builder request before it could run. Try a narrower prompt, or check the agent service configuration if this repeats.',
    tooLong: 'The prompt is too long. Shorten it to the key series, date range, mints, and completion goal.',
    submitFailed: 'The proposal request could not be submitted. Check AI provider and agent service configuration, then try again.',
  },
  agentRetry: 'Check the internal agent service configuration and retry.',
}
