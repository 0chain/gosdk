describe('allocation',   () => {
  beforeEach(()=>{
    cy.createWallet()
  })


  it('create allocation',()=>{
    cy.createAllocation()
  })

})