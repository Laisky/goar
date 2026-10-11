# Dependency and offline qualification

The consolidated change preserves both original PR histories (#1 and #2), updates go-ethereum to 1.17.8, and uses Go 1.26.9. The original Hamba decoder silently accepted truncated input and had a called vulnerability without a fixed version. Its replacement uses the existing maintained goavro byte codec, preserves the legacy negative-count sized block wire encoding, and rejects truncated or trailing bytes. Golden legacy bytes, both serializers, Unicode/empty fields, empty arrays, mixed positive/negative blocks, malformed lengths and an extreme negative count are covered.

GraphQL failure response bodies are treated as literal text, including percent signs. Controls cover ordinary errors, rate limits and successful responses. The missing JPEG fixture is replaced with deterministic temporary data at Merkle chunk boundaries; streaming and in-memory leaf sizes and hashes must agree.

Acceptance used Go 1.26.9: module verification, vet, all-package build, the full utils race suite, and a separate synthetic API/signature race suite. Both safe suites also ran in read-only networkless Docker containers. A generated ephemeral Ethereum key signs a bundle item; recovered address, decoded signature and tamper rejection are checked. No key file is read and no transaction is submitted. The source vulnerability scan found 0 called and 0 imported vulnerabilities, with 5 advisories in unused required-module packages. The module sums are now tracked for reproducibility.

The repository's wallet-funded root/example integration tests were not executed; they require financial transactions/private fixtures outside this qualification scope. No deployment was performed.
