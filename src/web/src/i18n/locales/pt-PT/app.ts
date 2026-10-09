import type { Messages } from '../../types'

const messages: Messages['app'] = {
  brandFull: 'Aurearia - Coleção de Moedas',
  brandTagline: 'Coleção de Moedas',
  header: {
    selectCoins: 'Selecionar moedas',
    cancelSelection: 'Cancelar modo de seleção',
  },
  sidebar: {
    reorderMenu: 'Reordenar menu',
    submenuLabel: 'Vistas de {label}',
  },
  openAgent: 'Abrir agente de IA',
  emailPrompt: {
    title: 'Adicione o seu email',
    body: 'Agora é obrigatório indicar um endereço de email. Adicione o seu para continuar a usar todas as funcionalidades.',
  },
  onboarding: {
    body: 'Comece pelo guia Primeiros passos para importar um CSV, experimentar a Captura rápida ou a Identificação de moedas e rever as definições que ativam alertas, imagens e ferramentas de IA.',
    openGuide: 'Abrir guia',
  },
}

export default messages
