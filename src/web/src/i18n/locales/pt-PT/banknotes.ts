import type { Messages } from '../../types'

const messages: Messages['banknotes'] = {
  title: 'Notas de banco',
  add: 'Adicionar nota',
  addTitle: 'Adicionar nota de banco',
  badge: 'Nota de banco',
  searchPlaceholder: 'Pesquisar notas por nome, país, inscrição…',
  failed: 'Não foi possível carregar as notas. Tente de novo daqui a pouco.',
  repeatedOnly: 'Só repetidas para troca',
  kpis: {
    label: 'Números das notas',
    banknotes: 'Notas',
    countries: 'Países',
    value: 'Valor total',
    unique: 'Notas únicas',
  },
  empty: {
    title: 'Ainda não tem notas de banco',
    body: 'Adicione uma nota, ou marque uma peça da coleção como nota de banco quando a editar.',
    noMatchTitle: 'Nenhuma nota corresponde',
    noMatchBody: 'Experimente outra pesquisa ou outro país.',
  },
}

export default messages
