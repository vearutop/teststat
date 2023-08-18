package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleRace = `
WARNING: DATA RACE
Read at 0x00c0006fd608 by goroutine 44:
  runtime.slicebytetostringtmp()
      runtime/string.go:154 +0x0
  github.com/some-lib/slowcache.(*bucket).Get()
      github.com/some-lib/slowcache@v1.5.1/slowcache.go:380 +0x496
  github.com/some-lib/slowcache.(*Cache).Has()
      github.com/some-lib/slowcache@v1.5.1/slowcache.go:169 +0xa4
  github.com/acme/foo/core/bytecache.(*ByteCache).delete()
      github.com/acme/foo/core/bytecache/byte_cache.go:106 +0xad
  github.com/acme/foo/core/bytecache.(*ByteCache).DeleteString()
      github.com/acme/foo/core/bytecache/byte_cache.go:90 +0xb1
  github.com/acme/foo/core/bytecache.(*invalidatingByteCache).SubscriberReceived()
      github.com/acme/foo/core/bytecache/invalidating_byte_cache.go:67 +0x81
  github.com/acme/foo/core/pubsub.(*MultiSubscriber).SubscriberReceived()
      github.com/acme/foo/core/pubsub/multi_subscriber.go:41 +0x16b
  github.com/acme/foo/core/pubsub.(*TestClient).publish.func1()
      github.com/acme/foo/core/pubsub/test_client.go:70 +0x73

Previous write at 0x00c0006fd60f by goroutine 37:
  runtime.slicecopy()
      runtime/slice.go:295 +0x0
  github.com/acme/foo/core/iron.encodeKey()
      github.com/acme/foo/core/iron/iron_bytecache.go:115 +0x313
  github.com/acme/foo/core/iron.(*ironByteCache).bufferKey()
      github.com/acme/foo/core/iron/iron_bytecache.go:101 +0x94
  github.com/acme/foo/core/iron.(*ironByteCache).Delete()
      github.com/acme/foo/core/iron/iron_bytecache.go:82 +0xdc
  github.com/acme/foo/core/trackerapi.(*ironClient).saveGenTracker()
      github.com/acme/foo/core/trackerapi/iron_client.go:244 +0x6c5
  github.com/acme/foo/core/trackerapi.(*ironClient).UpdateTracker()
      github.com/acme/foo/core/trackerapi/iron_client.go:185 +0x1b0
  github.com/acme/foo/core/trackerapi.(*ironClient).CreateTrackerAndLink()
      github.com/acme/foo/core/trackerapi/iron_client.go:76 +0x817
  github.com/acme/foo/core/trackerapi.(*Client).createTrackerIniron()
      github.com/acme/foo/core/trackerapi/client.go:593 +0xb71
  github.com/acme/foo/core/trackerapi.(*Client).findOrCreateTracker()
      github.com/acme/foo/core/trackerapi/client.go:369 +0x2df1
  github.com/acme/foo/core/trackerapi.(*Client).FindOrCreateTrackerTask()
      github.com/acme/foo/core/trackerapi/client.go:99 +0x16e
  github.com/acme/foo/core/trackerapi.(*TestClient).FindOrCreateTracker()
      github.com/acme/foo/core/trackerapi/test_client.go:70 +0x31d
  github.com/acme/foo/core/partner/partners/testsuite.(*PartnerSuite).setTrackerFixture()
      github.com/acme/foo/core/partner/partners/testsuite/partner_suite.go:1497 +0xd7
  github.com/acme/foo/core/partner/partners/testsuite.(*PartnerSuite).callbackDataFixture()
      github.com/acme/foo/core/partner/partners/testsuite/partner_suite.go:813 +0x2c04
  github.com/acme/foo/core/partner/partners/testsuite.(*PartnerSuite).TestAndroidGetSessionCallbacks()
      github.com/acme/foo/core/partner/partners/testsuite/partner_suite.go:1190 +0x30
  runtime.call16()
      runtime/asm_amd64.s:709 +0x48
  reflect.Value.Call()
      reflect/value.go:339 +0xd7
  github.com/stretchr/testify/suite.Run.func1()
      github.com/stretchr/testify@v1.7.0/suite/suite.go:158 +0x6dc
  testing.tRunner()
      testing/testing.go:1409 +0x213
  testing.(*T).Run.func1()
      testing/testing.go:1456 +0x47

Goroutine 44 (running) created at:
  github.com/acme/foo/core/pubsub.(*TestClient).publish()
      github.com/acme/foo/core/pubsub/test_client.go:68 +0x26c
  github.com/acme/foo/core/pubsub.(*TestClient).Publish()
      github.com/acme/foo/core/pubsub/test_client.go:42 +0xf2
  github.com/acme/foo/core/pubsub.(*Pubsub).Publish.func1()
      github.com/acme/foo/core/pubsub/pubsub.go:108 +0xb4
  github.com/acme/foo/core/pubsub.(*Pubsub).Publish.func2()
      github.com/acme/foo/core/pubsub/pubsub.go:113 +0x58

Goroutine 37 (running) created at:
  testing.(*T).Run()
      testing/testing.go:1456 +0x724
  github.com/stretchr/testify/suite.runTests()
      github.com/stretchr/testify@v1.7.0/suite/suite.go:203 +0x18f
  github.com/stretchr/testify/suite.Run()
      github.com/stretchr/testify@v1.7.0/suite/suite.go:176 +0x969
  github.com/acme/foo/core/partner/partners/analytics_test.TestAmobeeNoSessionSuite()
      github.com/acme/foo/core/partner/partners/analytics/amobee_test.go:120 +0x96f
  testing.tRunner()
      testing/testing.go:1409 +0x213
  testing.(*T).Run.func1()
      testing/testing.go:1456 +0x47


`

func Test_stripDataRace(t *testing.T) {
	b, err := json.MarshalIndent(stripDataRace(strings.Split(sampleRace, "\n")), "", " ")
	require.NoError(t, err)

	println(string(b))
}
