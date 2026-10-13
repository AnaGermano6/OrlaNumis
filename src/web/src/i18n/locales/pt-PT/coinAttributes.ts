import type { Messages } from '../../types'

const messages: Messages['coinAttributes'] = {
  presentation: {
    loose: 'Solta',
    coincard: 'Coincard',
    capsule: 'Cápsula',
    flip: 'Carteira / 2x2',
    proof_set: 'Estojo / proof',
    blister: 'Blister',
    slab: 'Certificada (NGC/PCGS)',
  },
  rarity: {
    common: 'Comum',
    uncommon: 'Pouco comum',
    scarce: 'Escassa',
    rare: 'Rara',
    very_rare: 'Muito rara',
    extremely_rare: 'Extremamente rara',
  },
  circulation: {
    current: 'Corrente',
    current_commemorative: 'Corrente · Comemorativa',
    demonetized: 'Desmonetizada',
    demonetized_commemorative: 'Desmonetizada · Comemorativa',
    non_circulating: 'Não circulante',
    non_circulating_commemorative: 'Não circulante · Comemorativa',
  },
}

export default messages
