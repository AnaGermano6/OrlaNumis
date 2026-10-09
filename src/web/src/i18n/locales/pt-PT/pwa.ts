import type { Messages } from '../../types'

const messages: Messages['pwa'] = {
  update: {
    title: 'Atualização disponível',
    body: 'Está disponível uma nova versão do Aurearia. Atualize para a instalar.',
  },
  install: {
    share: 'Partilhar',
    menu: 'menu',
    addToHomeScreen: 'Adicionar ao ecrã principal',
    addToPhone: 'Adicionar ao telemóvel',
    installApp: 'Instalar aplicação',
    install: 'Instalar',
    iosSafari: 'Toque no botão {share} {icon} e depois em {action}',
    iosEdge: 'Toque no botão {menu} {icon} e depois em {action}',
    iosOther: 'Para uma melhor experiência, abra no {safari}, toque em {icon} e depois em {action}',
    android: 'Toque no {menu} {icon} e depois em {action} ou {install}',
    otherBrowser: 'Use o menu do navegador para {install} ou {action}',
  },
}

export default messages
