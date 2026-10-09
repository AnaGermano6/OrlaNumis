import type { Messages } from '../../types'

const messages: Messages['availability'] = {
  title: 'Histórico de verificações de disponibilidade',
  back: 'Voltar à lista de desejos',
  signInTitle: 'É necessário iniciar sessão',
  signInBody: 'Inicie sessão para ver o histórico de verificações de disponibilidade.',
  runNotFound: 'Verificação não encontrada',
  runNumber: 'Verificação n.º {id}',
  unavailShort: '{count} indisp.',
  noResults: 'Sem resultados para esta verificação.',
  coin: 'Moeda',
  reason: 'Motivo',
  backToRuns: 'Voltar a todas as verificações',
  loadFailedTitle: 'Não foi possível carregar o histórico',
  empty: {
    title: 'Ainda não há verificações de disponibilidade',
    body: 'As verificações aparecem aqui sempre que a disponibilidade dos URLs da sua lista de desejos é verificada.',
  },
  page: 'Página {page}',
  loadFailed: 'Não foi possível carregar o histórico de verificações de disponibilidade. Tente novamente.',
  detailFailed: 'Não foi possível encontrar esta verificação ou já não está disponível.',
  runStatus: {
    done: 'concluída',
    queued: 'em fila',
    running: 'em execução',
    failed: 'falhou',
    partialFailure: 'falha parcial',
  },
  resultStatus: {
    available: 'disponível',
    unavailable: 'indisponível',
    unknown: 'desconhecido',
  },
  trigger: {
    scheduled: 'agendada',
    manual: 'manual',
  },
}

export default messages
