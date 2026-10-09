import type { Messages } from '../../types'

const messages: Messages['estimate'] = {
  status: 'Estimativa de valor {status}. Isto continuará em segundo plano; pode sair desta página.',
  queued: 'Estimativa de valor em fila. Pode sair desta página; será notificado quando terminar.',
  failed: 'Não foi possível estimar o valor',
  updateFailed: 'Não foi possível atualizar o valor da moeda',
  failedRetry: 'A estimativa de valor falhou. Tente novamente.',
  noResult: 'A IA não devolveu nenhuma estimativa',
  ready: 'Estimativa de valor pronta.',
  comparable: 'Comparável',
}

export default messages
