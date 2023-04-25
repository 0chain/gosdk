describe('wallet',   () => {
  it('Create wallet',async () => {
    cy.intercept('POST', 'http://127.0.0.1:8080/wallet').as('createWallet');
    cy.intercept('GET', 'https://dev.zus.network/dns/network').as('getDNS');

    cy.visit('http://127.0.0.1:8080')
    cy.wait('@getDNS').its('response.statusCode').should('eq', 200)

    cy.window().then(async win=>{
      const resp = await win.fetch('http://127.0.0.1:8080/wallet', {
        method: 'POST',
      })

      const w = await resp.json()
      const wallet = {} 
      wallet.client_id = w.client_id
      wallet.mnemonics = w.mnemonics
      w.keys.forEach(it=>{
        wallet.public_key = it.public_key
        wallet.private_key = it.private_key
      })

      cy.get('#clientId').invoke('val', wallet.client_id)
      cy.get('#publicKey').invoke('val', wallet.public_key)
      cy.get('#privateKey').invoke('val', wallet.private_key)
      cy.get('#mnemonic').invoke('val',wallet.mnemonics)
    })


    cy.wait('@createWallet').its('response.statusCode').should('eq', 200)
    cy.wait(1500)

    cy.get('#btnShowLogs').click()
    cy.get('#btnSetWallet').click()
    cy.wait(1500)
    cy.get("#btnSendMeTokens").click()

  })

})