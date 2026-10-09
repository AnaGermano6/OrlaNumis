import type { Messages } from '../../types'

const messages: Messages['editCoin'] = {
  title: 'Editar moeda',
  loadFailed: 'Não foi possível carregar a moeda',
  cleanupPending: 'As suas alterações foram guardadas. Os ficheiros de imagem antigos aguardam limpeza; a API tentará novamente no arranque. Não volte a carregar a imagem de substituição.',
  cleanupPendingTitle: 'Limpeza de imagens pendente',
  imageNotUploaded: 'Os detalhes da moeda foram guardados, mas não foi possível carregar a imagem. Tente adicionar a imagem novamente.',
  updateFailed: 'Não foi possível atualizar a moeda',
}

export default messages
