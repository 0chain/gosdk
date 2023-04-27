// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//


Cypress.Commands.add('createWallet', () => { 

  const n = Cypress.env("NETWORK_URL") || "dev.zus.network"

  cy.log("network: ",n)

  cy.intercept('POST', 'http://127.0.0.1:8080/wallet').as('createWallet');
  cy.intercept('GET', 'https://'+n+'/dns/network').as('getDNS');
  cy.intercept('GET','**/v1/block/get?round=*').as('faucet')
  cy.intercept('GET','**/v1/client/get/balance?client_id=*').as("getBalance")

  cy.visit('http://127.0.0.1:8080?network='+n)
  cy.wait('@getDNS').its('response.statusCode').should('eq', 200)

  cy.wait(1500)

  cy.get('#btnShowLogs').click()

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
    cy.get('#btnSetWallet').click()
  })


  cy.wait('@createWallet').its('response.statusCode').should('eq', 200)
  cy.wait(1500)

  cy.get("#btnSendMeTokens").click()
  cy.wait(3000)
  //retry one time
  cy.get("#btnSendMeTokens").click()
  cy.wait("@faucet", {timeout:60*1000}).its('response.statusCode').should('eq', 200)


  cy.get("#btnGetBalance").click()

  const waitGetBalance = ()=> {
    cy.wait("@getBalance").then(it=>{
      if (it.response.statusCode == 200) {
        expect(it.response.body.balance).to.greaterThan(9000000000)
      }else{
        waitGetBalance()
      }
    })
  }

  waitGetBalance()
 })

 Cypress.Commands.add('createAllocation', () => { 
   cy.get('#btnCreateAllocation').click()
 })
