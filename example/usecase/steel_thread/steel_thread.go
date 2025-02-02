// Annotated steel thread flow for calling out all signing, verification, and key management.
// This example is focused on the signing, exchange, and verification of objects.

// We assume there are two parties:
//	1 . Alice, using a wallet, applying for a credential
//	2. An issuer via the SSI Service, processing credential applications and issuing credentials

// Each party has a single DID. Alice and the SSI Service each have a single did:key DID.
// Alice stores her DID’s private key in her wallet.
// The SSI Service stores its private key in the service key store database.

package main

import (
	"fmt"
	"github.com/TBD54566975/ssi-sdk/credential"
	"github.com/TBD54566975/ssi-sdk/credential/exchange"
	"github.com/TBD54566975/ssi-sdk/credential/manifest"
	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/example"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
)

type Entity struct {
	didKey                did.DIDKey
	verifier              crypto.JWTVerifier
	credentialManifest    manifest.CredentialManifest
	credentialApplication manifest.CredentialApplication
	credentialResponse    manifest.CredentialResponse
	verifiableCredentials []credential.VerifiableCredential
}

func (t *Entity) GenerateWallet() {
	walletDIDPrivateKey, walletDIDKey, err := did.GenerateDIDKey(crypto.Ed25519)
	example.HandleExampleError(err, "Failed to generate DID")
	walletDIDWJWK, err := jwk.New(walletDIDPrivateKey)
	example.HandleExampleError(err, "Failed to generate JWK")
	walletSigner, err := crypto.NewJWTSigner(walletDIDKey.ToString(), walletDIDWJWK)
	example.HandleExampleError(err, "Failed to generate signer")
	walletVerifier, err := walletSigner.ToVerifier()
	example.HandleExampleError(err, "Failed to generate verifier")

	t.didKey = *walletDIDKey
	t.verifier = *walletVerifier
}

func (t *Entity) CreateCredentialManifest() {
	credManifest := createCredentialManifest(t.didKey.ToString())
	t.credentialManifest = credManifest
}

func (t *Entity) CreateCredentialApplication() {
	credApplication := createCredentialApplication(t.credentialManifest)
	t.credentialApplication = credApplication
}

func (t *Entity) SignCredentialManifest() {
	//TODO: SignCredentialManifestJWT(issuer) -> JWTString
}

func (t *Entity) SignCredentialApplication() {
	//TODO: SignCredentialApplicationJWT(issuer) -> JWTString
}

func (t *Entity) SignCredentialResponse() {
	//TODO: SignCredentialResponseJWT(issuer) -> JWTString
}

func (t *Entity) SetCredentialManifest(manifest manifest.CredentialManifest) {
	t.credentialManifest = manifest
}

func (t *Entity) SetCredentialApplication(application manifest.CredentialApplication) {
	t.credentialApplication = application
}

func (t *Entity) SetCredentialResponse(response manifest.CredentialResponse) {
	t.credentialResponse = response
}

func (t *Entity) SetVerifiableCredentials(credentials []credential.VerifiableCredential) {
	t.verifiableCredentials = credentials
}

func (t *Entity) ValidateCredentialManifest() error {
	//TODO: Validate Signature
	return t.credentialManifest.IsValid()
}

func (t *Entity) ValidateCredentialApplication() error {
	//TODO: Validate Signature
	return t.credentialApplication.IsValid()
}

func (t *Entity) ValidateCredentialResponse() error {
	//TODO: Validate Signature
	return t.credentialResponse.IsValid()
}

func (t *Entity) ValidateVerifiableCredentials() error {

	for _, vc := range t.verifiableCredentials {
		if vc.IsValid() != nil {
			return vc.IsValid()
		}
	}
	return nil
}

func (t *Entity) ProcessCredentialApplication(issuer string, subject string) (*manifest.CredentialResponse, []credential.VerifiableCredential) {

	var creds []credential.VerifiableCredential
	for _, od := range t.credentialManifest.OutputDescriptors {
		// TODO: Create Cred off of OD
		creds = append(creds, createVerifiableCredential(issuer, subject, od))
	}

	responseBuilder := manifest.NewCredentialResponseBuilder(t.credentialManifest.ID)

	if err := responseBuilder.SetApplicationID(t.credentialApplication.ID); err != nil {
		example.HandleExampleError(err, "could not fulfill credential application: could not set application id")
	}

	/**
		 The SSI Service creates a Credential Response, CR, signed with didI
	**/

	var descriptors []exchange.SubmissionDescriptor
	for i, c := range creds {

		format := string(exchange.JWTVC)

		descriptors = append(descriptors, exchange.SubmissionDescriptor{
			ID:     c.ID,
			Format: format,
			Path:   fmt.Sprintf("$.verifiableCredential[%d]", i),
		})
	}

	// set the information for the fulfilled credentials in the response
	if err := responseBuilder.SetFulfillment(descriptors); err != nil {
		example.HandleExampleError(err, "could not fulfill credential application: could not set fulfillment")
	}
	credRes, err := responseBuilder.Build()
	if err != nil {
		example.HandleExampleError(err, "could not build response")
	}

	return credRes, creds
}

func (t *Entity) FlexFullyValidatedCredentials() {
	fmt.Print("\n#ShowingOffMyNewlyMintedCredentials: \n")
	for _, vc := range t.verifiableCredentials {
		fmt.Print(vc.CredentialSubject)
	}
}

func main() {

	/**
		Step 1: Alice creates a DID, stored in her wallet, didW
	**/
	aliceWalletEntity := new(Entity)
	aliceWalletEntity.GenerateWallet()

	/**
		Step 2: The issuer creates a DID, stored in the SSI Service, didI
	**/
	issuerWalletEntity := new(Entity)
	issuerWalletEntity.GenerateWallet()

	/**
		Step 3: The Credential Manifest CM is created and signed by the Issuer, with didI
	**/
	issuerWalletEntity.CreateCredentialManifest()
	issuerWalletEntity.SignCredentialManifest()

	/**
		Step 4: Alice fetches `Credential Manifest` and…
			* Validates the signature using `didI`
			* Verifies it is a valid and well-formed Credential Manifest
	**/
	aliceWalletEntity.SetCredentialManifest(issuerWalletEntity.credentialManifest)
	if err := aliceWalletEntity.ValidateCredentialManifest(); err != nil {
		example.HandleExampleError(err, "could not verify credential manifest")
	}

	/**
		Step 5: Alice forms a response to Credential Manifest, Creates a Credential Application
	**/
	aliceWalletEntity.CreateCredentialApplication()
	aliceWalletEntity.SignCredentialApplication()

	/**
		Step 6: Alice submits `Credential Application` to the SSI Service. The SSI Service processes `Credential Application`...
			* Validates the signature using `didW`
			* Verifies it is a valid and well-formed Credential Application
			* Validates it complies with `CM`
	**/
	issuerWalletEntity.SetCredentialApplication(aliceWalletEntity.credentialApplication)
	if err := issuerWalletEntity.ValidateCredentialApplication(); err != nil {
		example.HandleExampleError(err, "could not verify credential application")
	}

	/**
		Step 7: The SSI Service creates Verifiable Credentials signed with `didI`. The SSI Service creates a
		Credential Response signed with `didI`
	**/
	credentialResponse, verifiableCredentials := issuerWalletEntity.ProcessCredentialApplication(issuerWalletEntity.didKey.ToString(), aliceWalletEntity.didKey.ToString())

	/**
		Step 8: Alice receives Credential Response containing the Verifiable Credentials and
			* Validates the signature of the `CR` using `didI`
			* Verifies `CR` and `VC` are well-formed
			* Verifies the signature(s) of `VC` using `didI`
			* Possible other validity checks, such as making sure the Credential is for Alice and has expected data

	**/
	aliceWalletEntity.SetCredentialResponse(*credentialResponse)
	if err := aliceWalletEntity.ValidateCredentialResponse(); err != nil {
		example.HandleExampleError(err, "could not verify credential response")
	}

	aliceWalletEntity.SetVerifiableCredentials(verifiableCredentials)
	if err := aliceWalletEntity.ValidateCredentialResponse(); err != nil {
		example.HandleExampleError(err, "could not verify verifiable credentials")
	}

	/**
		Step 8: Alice flexes her newly minted credentials to her friends and family
	**/
	aliceWalletEntity.FlexFullyValidatedCredentials()
}

func createCredentialApplication(cm manifest.CredentialManifest) manifest.CredentialApplication {
	return manifest.CredentialApplication{
		ID:          uuid.New().String(),
		SpecVersion: "https://identity.foundation/credential-manifest/spec/v1.0.0/",
		ManifestID:  cm.ID,
		Format: &exchange.ClaimFormat{
			JWT: &exchange.JWTType{Alg: []crypto.SignatureAlgorithm{crypto.EdDSA}},
		},
		PresentationSubmission: &exchange.PresentationSubmission{
			ID:           "psid",
			DefinitionID: "definitionId",
			DescriptorMap: []exchange.SubmissionDescriptor{
				{
					ID:     "ps-id",
					Format: "jwt",
					Path:   "path",
				},
			},
		},
	}
}

func createVerifiableCredential(issuerDID string, walletDID string, descriptor manifest.OutputDescriptor) credential.VerifiableCredential {
	knownIssuanceDate := "2020-01-01T19:23:24Z"
	knownSubject := map[string]interface{}{
		"id":        string(walletDID),
		"birthdate": "1975-01-01",
	}

	vcBuilder := credential.NewVerifiableCredentialBuilder()

	err := vcBuilder.SetIssuer(issuerDID)
	example.HandleExampleError(err, "Failed to set issuer")
	err = vcBuilder.SetIssuanceDate(knownIssuanceDate)
	example.HandleExampleError(err, "Failed to set issuance date")
	err = vcBuilder.SetCredentialSubject(knownSubject)
	example.HandleExampleError(err, "Failed to set subject")

	vc, err := vcBuilder.Build()
	example.HandleExampleError(err, "Failed to make verifiable credential")
	example.HandleExampleError(vc.IsValid(), "Verifiable credential is not valid")

	return *vc
}

func createCredentialManifest(issuer string) manifest.CredentialManifest {
	return manifest.CredentialManifest{
		ID:          "WA-DL-CLASS-A",
		SpecVersion: "https://identity.foundation/credential-manifest/spec/v1.0.0/",
		Issuer: manifest.Issuer{
			ID: issuer,
		},
		PresentationDefinition: &exchange.PresentationDefinition{
			ID: "pres-def-id",
			InputDescriptors: []exchange.InputDescriptor{
				{
					ID: "test-id",
					Constraints: &exchange.Constraints{
						Fields: []exchange.Field{
							{
								Path: []string{".vc.id"},
							},
						},
					},
				},
			},
		},
		OutputDescriptors: []manifest.OutputDescriptor{
			{
				ID:          "id1",
				Schema:      "https://test.com/schema",
				Name:        "good ID",
				Description: "it's all good",
			},
			{
				ID:          "id2",
				Schema:      "https://test.com/schema",
				Name:        "good ID",
				Description: "it's all good",
			},
		},
	}
}
