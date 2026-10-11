# Maintained implementation provenance

This initial release extracts the maintained Tile client from the Port OS backend's `integrations/go-tile` module. The source repository was at commit `8863bb707` with the maintained Go 1.25 dependency update in its working tree. No live credentials or device captures were used to publish this release.

The research and replay strategy documents were copied with the implementation. Their historical `integrations/go-tile` paths describe the original capture location; this repository owns the corresponding `test/fixtures` paths. Replay envelopes are synthetic or redacted reference fixtures as documented, not newly captured production traffic.

Before publication, fixture tests, examples, build, vet and Staticcheck passed. Test-only helpers were moved into a `_test.go` file, and CI now enforces formatting, build, vet, lint and fixture tests.
