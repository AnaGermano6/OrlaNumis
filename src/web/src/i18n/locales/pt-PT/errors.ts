import type { Messages } from '../../types'

const messages: Messages['errors'] = {
  agentUnavailable: 'Serviço do agente indisponível. Verifique a configuração do serviço interno do agente.',
  agentCredentialMissing: 'A credencial do serviço interno do agente não está configurada. Verifique a configuração do serviço interno do agente.',
  agentCheckConfig: 'Verifique a configuração do serviço do agente e tente novamente.',
  agentRetry: 'Verifique a configuração do serviço interno do agente e tente novamente.',
}

export default messages
