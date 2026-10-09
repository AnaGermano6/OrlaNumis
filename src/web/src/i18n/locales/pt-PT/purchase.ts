import type { Messages } from '../../types'

const messages: Messages['purchase'] = {
  purchasedFrom: 'Comprada a',
  purchasedFromPlaceholder: 'ex.: VCoins, Heritage Auctions',
  allOptional: 'Todos os campos são opcionais. Pode atualizá-los mais tarde.',
  addShipment: 'Adicionar seguimento do envio agora',
  shipmentHint: 'Introduza apenas o número de seguimento. O ParcelApp usará o título da moeda como descrição da entrega.',
  shipmentNotes: 'Notas do envio (opcional)',
  trackingRequired: 'O número de seguimento é obrigatório quando o seguimento do envio está ativo',
}

export default messages
