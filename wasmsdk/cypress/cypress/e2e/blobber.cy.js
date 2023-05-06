describe('blobber',   () => {
  beforeEach(()=>{
    cy.createWallet()
    cy.createAllocation()
  })


  it('upload and download file',()=>{
    cy.get('input[type=file]').selectFile({
      contents: Cypress.Buffer.from('test file content'),
      fileName: 'file.txt',
      lastModified: Date.now(),
    })
  })

})