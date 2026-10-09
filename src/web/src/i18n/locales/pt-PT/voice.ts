import type { Messages } from '../../types'

const messages: Messages['voice'] = {
  privacy: 'O reconhecimento de voz pode enviar o áudio para o fornecedor do navegador ou do dispositivo. O Aurearia não guarda áudio.',
  start: 'Iniciar ditado por voz',
  stop: 'Parar ditado por voz',
  requesting: 'A pedir acesso ao microfone',
  listening: 'A ouvir',
  stopping: 'A parar',
  dictate: 'Ditar notas',
  errors: {
    blocked: 'O acesso ao microfone está bloqueado. Autorize-o nas permissões do site no navegador e tente novamente.',
    noSpeech: 'Não foi detetada voz. Tente novamente ou continue a escrever.',
    noMic: 'Não há microfone disponível. Pode continuar a escrever as notas.',
    network: 'O reconhecimento de voz está indisponível. Verifique a ligação ou continue a escrever.',
    stopped: 'O ditado por voz parou inesperadamente. Pode continuar a escrever as notas.',
    inUse: 'O microfone está a ser usado por outra aplicação ou não foi possível iniciá-lo.',
    requestFailed: 'Não foi possível pedir acesso ao microfone. Pode continuar a escrever as notas.',
    startFailed: 'Não foi possível iniciar o ditado por voz. Pode continuar a escrever as notas.',
  },
}

export default messages
