describe('wallet',   () => {

  const wallet = {}
  
  it("create wallet", ()=>{
     cy.request('POST','http://127.0.0.1:8080/wallet').then(response =>{
      const w = JSON.parse(response.body)
      wallet.client_id = w.client_id
      wallet.mnemonics = w.mnemonics
      w.keys.forEach(it=>{
        wallet.public_key = it.public_key
        wallet.private_key = it.private_key
      }) 
    })
  })

  it('Set wallet',async () => {
    await cy.visit('http://127.0.0.1:8080')


    cy.log("wallet:",wallet);
  })

})