import type { Messages } from '../../types'

const messages: Messages['health'] = {
  score: 'Pontuação de saúde',
  allPassed: 'Todas as verificações de qualidade foram aprovadas',
  items: {
    denomination: 'Valor facial',
    country: 'País',
    monetarySystem: 'Sistema monetário',
    mint: 'Casa da moeda',
    category: 'Categoria',
    material: 'Material',
    grade: 'Grau',
    anyImage: 'Pelo menos uma fotografia',
    bothImages: 'Fotografias dos dois lados',
    value: 'Valor atual',
    recentValue: 'Um valor revisto nos últimos seis meses',
    aiAnalysis: 'Análise de IA dos dois lados',
    aiCoverage: 'Análise de IA do outro lado',
  },
  dimensions: {
    metadata: 'Metadados',
    images: 'Imagens',
    valuation: 'Avaliação',
  },
}

export default messages
