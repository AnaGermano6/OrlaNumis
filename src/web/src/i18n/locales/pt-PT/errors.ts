import type { Messages } from '../../types'

const messages: Messages['errors'] = {
  agentUnavailable: 'Serviço do agente indisponível. Verifique a configuração do serviço interno do agente.',
  agentCredentialMissing: 'A credencial do serviço interno do agente não está configurada. Verifique a configuração do serviço interno do agente.',
  agentCheckConfig: 'Verifique a configuração do serviço do agente e tente novamente.',
  setBuilder: {
    slotLimit: 'O pedido de criação do conjunto excedeu o limite de posições do fluxo. Experimente um pedido mais restrito por intervalo de datas, casa da moeda, governante ou série e volte a submeter.',
    rejected: 'O agente rejeitou o pedido de criação do conjunto antes de o executar. Experimente um pedido mais restrito ou, se isto se repetir, verifique a configuração do serviço do agente.',
    tooLong: 'O pedido é demasiado longo. Reduza-o à série principal, ao intervalo de datas, às casas da moeda e ao objetivo de conclusão.',
    submitFailed: 'Não foi possível submeter o pedido de proposta. Verifique a configuração do fornecedor de IA e do serviço do agente e tente novamente.',
  },
  agentRetry: 'Verifique a configuração do serviço interno do agente e tente novamente.',
}

export default messages
