import type { Messages } from '../../types'

const messages: Messages['coinCard'] = {
  unavailable: 'Indisponível',
  available: 'Disponível',
  unknown: 'Desconhecido',
  sold: 'Vendida: {amount}',
  paid: 'Pago: {amount}',
  soldTo: 'Para: {name}',
  purchased: 'Comprada',
}

export default messages
