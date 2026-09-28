// Package router is everything the router defines, as opposed to the
// providers that plug into it (package providers/...).
//
// It holds the two contracts and the machinery between them:
//
//   - Provider (provider.go) is what every backend is: a name, a provider
//     type and Verify. What a backend can do is expressed by the use-case
//     interfaces it also implements (chat.Provider, embedding.Provider, ...).
//   - UseCase (usecase.go) is what the server serves. Each use case lives in
//     a sub-package (router/chat, router/embedding, ...) that owns its
//     contract end to end: its wire types (api.go), its model type, the
//     provider interface a backend implements to serve it, its request
//     handler, and its test scenarios (scenarios.go). A use case declares
//     itself with a Spec and embeds the Base built from it, which implements
//     everything in UseCase except Handle.
//   - The machinery every use case shares, written once: the model catalog
//     (catalog.go), model-string resolution (resolve.go) and the scenario test
//     runner (test.go).
//
// Model lists are fetched once, when a use case is built at startup, and never
// change afterwards; restarting the server is how it picks up new models.
package router
