export default {
  checkAvailability: 'Check Availability',
  searchAlerts: 'Search Alerts',
  runHistory: 'Run History',
  addTo: 'Add to wishlist',
  finderAgent: 'Add Wish List Finder Agent',
  check: {
    available: '{count} available',
    unavailable: '{count} unavailable',
    unknown: '{count} unknown',
    checked: '{count} checked',
  },
  empty: {
    title: 'Your wishlist is empty',
    body: 'Add coins to your wishlist to track what you\'re looking for',
    searchAi: 'Search for Coins with AI',
  },
  add: {
    fromUrl: 'Add from URL',
    fromUrlHint: 'Import and review a public dealer listing.',
    fromImage: 'Add from image analysis',
    fromImageHint: 'Photograph or upload a coin for identification.',
    another: 'Add another way',
  },
  leave: {
    title: 'On your wishlist',
    message: {
      one: 'This coin was on your wishlist: {names}. Take it off the wishlist now that you have it?',
      other: 'These {count} wishlist coins are this one: {names}. Take them off the wishlist now that you have it?',
    },
    confirm: { one: 'Take it off', other: 'Take them off' },
    keep: 'Keep on the wishlist',
    done: { one: 'Taken off the wishlist.', other: '{count} coins taken off the wishlist.' },
    failed: 'The coin could not be taken off the wishlist. Delete it from its page on the wishlist.',
  },
}
