import type { Messages } from '../../types'

const messages: Messages['map'] = {
  title: 'Mapa das moedas',
  loading: 'A carregar as casas da moeda da coleção...',
  unavailable: 'Mapa das casas da moeda indisponível',
  noLocations: 'Não há casas da moeda configuradas',
  noLocationsBody: 'Peça a um administrador que adicione casas da moeda globais antes de usar o mapa.',
  noCoins: 'Não há moedas ativas para mapear',
  noCoinsBody: 'Adicione moedas com casa da moeda ou volte depois de a sua coleção carregar.',
  summary: 'Resumo do mapa das casas da moeda',
  mappedCoins: 'Moedas mapeadas:',
  unknownMint: 'Casa da moeda desconhecida',
  locationsFailed: 'Não foi possível carregar as casas da moeda. Verifique a ligação e tente novamente.',
  collectionFailed: 'Não foi possível carregar a coleção ativa. Verifique a ligação e tente novamente.',
  pinLabel: { one: '{name}: {count} moeda', other: '{name}: {count} moedas' },
  selectedMint: 'Casa da moeda selecionada',
  inThisView: { one: '{count} moeda nesta vista', other: '{count} moedas nesta vista' },
  closeDrawer: 'Fechar painel da casa da moeda',
  unknownRuler: 'Governante desconhecido',
  unknownDenomination: 'Denominação desconhecida',
  unattributed: 'Casas da moeda não atribuídas',
  needReview: { one: '{count} moeda precisa de revisão da casa da moeda', other: '{count} moedas precisam de revisão da casa da moeda' },
  unknownMintHeading: 'Casa da moeda desconhecida',
  noCoordinate: 'Nenhuma coordenada de casa da moeda corresponde a este nome.',
  listLabel: 'Lista de casas da moeda',
  locations: 'Localizações',
  mapLabel: 'Mapa das casas da moeda',
}

export default messages
