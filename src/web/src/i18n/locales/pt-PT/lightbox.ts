import type { Messages } from '../../types'

const messages: Messages['lightbox'] = {
  viewer: 'Visualizador de imagem: {type}',
  closeEsc: 'Fechar (Esc)',
  cleanupPending: 'Imagem guardada. Os ficheiros de imagem antigos aguardam limpeza; a API tentará novamente no arranque. Não volte a carregar a imagem de substituição.',
  saveFailed: 'Não foi possível guardar a imagem. Tente novamente.',
}

export default messages
