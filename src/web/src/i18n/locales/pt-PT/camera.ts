import type { Messages } from '../../types'

const messages: Messages['camera'] = {
  focusOne: 'Enquadre uma moeda no círculo',
  capturePhoto: 'Tirar fotografia',
  starting: 'A iniciar a câmara...',
  capture: 'Capturar',
  startWhenReady: 'Inicie a câmara quando estiver pronto.',
  start: 'Iniciar câmara',
  use: 'Usar câmara',
  takePhoto: 'Tirar fotografia',
  uploadImage: 'Carregar imagem',
  capturePhotoAria: 'Capturar fotografia',
  uploadFromLibrary: 'Carregar da galeria',
  errors: {
    unavailableDevice: 'O acesso à câmara não está disponível neste dispositivo.',
    blocked: 'O acesso à câmara está bloqueado. Ative-o nas definições do navegador ou do site.',
    denied: 'A permissão da câmara foi recusada.',
    deniedCanUpload: 'A permissão da câmara foi recusada. Pode continuar a carregar imagens.',
    notFound: 'Não foi encontrada nenhuma câmara neste dispositivo.',
    unavailable: 'A câmara não está disponível.',
    unavailableCanUpload: 'A câmara não está disponível. Pode continuar a carregar imagens.',
    notReady: 'A câmara ainda não está pronta. Tente novamente dentro de momentos.',
    displaySize: 'Não foi possível determinar o tamanho de apresentação do vídeo.',
    prepareCanUpload: 'Não foi possível preparar a captura de imagem. Pode continuar a carregar imagens.',
    captureFailed: 'Não foi possível capturar a imagem da câmara.',
    captureFailedCanUpload: 'Não foi possível capturar a imagem da câmara. Pode continuar a carregar imagens.',
  },
}

export default messages
