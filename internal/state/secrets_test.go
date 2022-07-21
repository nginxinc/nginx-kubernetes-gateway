// nolint:gosec
package state_test

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/statefakes"
)

var (
	cert = []byte(`-----BEGIN CERTIFICATE-----
MIIDLjCCAhYCCQDAOF9tLsaXWjANBgkqhkiG9w0BAQsFADBaMQswCQYDVQQGEwJV
UzELMAkGA1UECAwCQ0ExITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0
ZDEbMBkGA1UEAwwSY2FmZS5leGFtcGxlLmNvbSAgMB4XDTE4MDkxMjE2MTUzNVoX
DTIzMDkxMTE2MTUzNVowWDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxGTAXBgNVBAMMEGNhZmUuZXhh
bXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCp6Kn7sy81
p0juJ/cyk+vCAmlsfjtFM2muZNK0KtecqG2fjWQb55xQ1YFA2XOSwHAYvSdwI2jZ
ruW8qXXCL2rb4CZCFxwpVECrcxdjm3teViRXVsYImmJHPPSyQgpiobs9x7DlLc6I
BA0ZjUOyl0PqG9SJexMV73WIIa5rDVSF2r4kSkbAj4Dcj7LXeFlVXH2I5XwXCptC
n67JCg42f+k8wgzcRVp8XZkZWZVjwq9RUKDXmFB2YyN1XEWdZ0ewRuKYUJlsm692
skOrKQj0vkoPn41EE/+TaVEpqLTRoUY3rzg7DkdzfdBizFO2dsPNFx2CW0jXkNLv
Ko25CZrOhXAHAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKHFCcyOjZvoHswUBMdL
RdHIb383pWFynZq/LuUovsVA58B0Cg7BEfy5vWVVrq5RIkv4lZ81N29x21d1JH6r
jSnQx+DXCO/TJEV5lSCUpIGzEUYaUPgRyjsM/NUdCJ8uHVhZJ+S6FA+CnOD9rn2i
ZBePCI5rHwEXwnnl8ywij3vvQ5zHIuyBglWr/Qyui9fjPpwWUvUm4nv5SMG9zCV7
PpuwvuatqjO1208BjfE/cZHIg8Hw9mvW9x9C+IQMIMDE7b/g6OcK7LGTLwlFxvA8
7WjEequnayIphMhKRXVf1N349eN98Ez38fOTHTPbdJjFA/PcC+Gyme+iGt5OQdFh
yRE=
-----END CERTIFICATE-----`)
	key = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAqeip+7MvNadI7if3MpPrwgJpbH47RTNprmTStCrXnKhtn41k
G+ecUNWBQNlzksBwGL0ncCNo2a7lvKl1wi9q2+AmQhccKVRAq3MXY5t7XlYkV1bG
CJpiRzz0skIKYqG7Pcew5S3OiAQNGY1DspdD6hvUiXsTFe91iCGuaw1Uhdq+JEpG
wI+A3I+y13hZVVx9iOV8FwqbQp+uyQoONn/pPMIM3EVafF2ZGVmVY8KvUVCg15hQ
dmMjdVxFnWdHsEbimFCZbJuvdrJDqykI9L5KD5+NRBP/k2lRKai00aFGN684Ow5H
c33QYsxTtnbDzRcdgltI15DS7yqNuQmazoVwBwIDAQABAoIBAQCPSdSYnQtSPyql
FfVFpTOsoOYRhf8sI+ibFxIOuRauWehhJxdm5RORpAzmCLyL5VhjtJme223gLrw2
N99EjUKb/VOmZuDsBc6oCF6QNR58dz8cnORTewcotsJR1pn1hhlnR5HqJJBJask1
ZEnUQfcXZrL94lo9JH3E+Uqjo1FFs8xxE8woPBqjZsV7pRUZgC3LhxnwLSExyFo4
cxb9SOG5OmAJozStFoQ2GJOes8rJ5qfdvytgg9xbLaQL/x0kpQ62BoFMBDdqOePW
KfP5zZ6/07/vpj48yA1Q32PzobubsBLd3Kcn32jfm1E7prtWl+JeOFiOznBQFJbN
4qPVRz5hAoGBANtWyxhNCSLu4P+XgKyckljJ6F5668fNj5CzgFRqJ09zn0TlsNro
FTLZcxDqnR3HPYM42JERh2J/qDFZynRQo3cg3oeivUdBVGY8+FI1W0qdub/L9+yu
edOZTQ5XmGGp6r6jexymcJim/OsB3ZnYOpOrlD7SPmBvzNLk4MF6gxbXAoGBAMZO
0p6HbBmcP0tjFXfcKE77ImLm0sAG4uHoUx0ePj/2qrnTnOBBNE4MvgDuTJzy+caU
k8RqmdHCbHzTe6fzYq/9it8sZ77KVN1qkbIcuc+RTxA9nNh1TjsRne74Z0j1FCLk
hHcqH0ri7PYSKHTE8FvFCxZYdbuB84CmZihvxbpRAoGAIbjqaMYPTYuklCda5S79
YSFJ1JzZe1Kja//tDw1zFcgVCKa31jAwciz0f/lSRq3HS1GGGmezhPVTiqLfeZqc
R0iKbhgbOcVVkJJ3K0yAyKwPTumxKHZ6zImZS0c0am+RY9YGq5T7YrzpzcfvpiOU
ffe3RyFT7cfCmfoOhDCtzukCgYB30oLC1RLFOrqn43vCS51zc5zoY44uBzspwwYN
TwvP/ExWMf3VJrDjBCH+T/6sysePbJEImlzM+IwytFpANfiIXEt/48Xf60Nx8gWM
uHyxZZx/NKtDw0V8vX1POnq2A5eiKa+8jRARYKJLYNdfDuwolxvG6bZhkPi/4EtT
3Y18sQKBgHtKbk+7lNJVeswXE5cUG6EDUsDe/2Ua7fXp7FcjqBEoap1LSw+6TXp0
ZgrmKE8ARzM47+EJHUviiq/nupE15g0kJW3syhpU9zZLO7ltB0KIkO9ZRcmUjo8Q
cpLlHMAqbLJ8WYGJCkhiWxyal6hYTyWY4cVkC0xtTl/hUE9IeNKo
-----END RSA PRIVATE KEY-----`)

	invalidCert = []byte(`-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`)

	invalidKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----`)
)

var (
	secret1 = &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "secret1",
		},
		Data: map[string][]byte{
			apiv1.TLSCertKey:       cert,
			apiv1.TLSPrivateKeyKey: key,
		},
		Type: apiv1.SecretTypeTLS,
	}

	secret2 = &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "secret2",
		},
		Data: map[string][]byte{
			apiv1.TLSCertKey:       cert,
			apiv1.TLSPrivateKeyKey: key,
		},
		Type: apiv1.SecretTypeTLS,
	}

	secret3 = &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "secret3",
		},
		Data: map[string][]byte{
			apiv1.TLSCertKey:       cert,
			apiv1.TLSPrivateKeyKey: key,
		},
		Type: apiv1.SecretTypeTLS,
	}

	invalidSecretType = &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "invalid-type",
		},
		Data: map[string][]byte{
			apiv1.TLSCertKey:       cert,
			apiv1.TLSPrivateKeyKey: key,
		},
		Type: apiv1.SecretTypeDockercfg,
	}
	invalidSecretKey = &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "invalid-key",
		},
		Data: map[string][]byte{
			apiv1.TLSCertKey:       cert,
			apiv1.TLSPrivateKeyKey: invalidKey,
		},
		Type: apiv1.SecretTypeTLS,
	}
)

var _ = Describe("SecretDiskMemoryManager", func() {
	var (
		fakeStore     *statefakes.FakeSecretStore
		memMgr        state.SecretDiskMemoryManager
		tmpSecretsDir string
	)

	BeforeEach(OncePerOrdered, func() {
		dir, err := os.MkdirTemp("", "secrets-test")
		tmpSecretsDir = dir
		Expect(err).ToNot(HaveOccurred(), "failed to create temp directory for tests")

		fakeStore = &statefakes.FakeSecretStore{}
		memMgr = state.NewSecretDiskMemoryManager(tmpSecretsDir, fakeStore)
	})

	AfterEach(OncePerOrdered, func() {
		Expect(os.RemoveAll(tmpSecretsDir)).To(Succeed())
	})

	Describe("Manages secrets on disk", Ordered, func() {
		testRequest := func(s *apiv1.Secret, expPath string, expErr bool) {
			nsname := types.NamespacedName{Namespace: s.Namespace, Name: s.Name}
			actualPath, err := memMgr.Request(nsname)

			if expErr {
				Expect(err).To(HaveOccurred())
				Expect(actualPath).To(BeEmpty())
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(actualPath).To(Equal(expPath))
			}
		}

		It("should return an error and empty path when secret does not exist", func() {
			fakeStore.GetReturns(nil)

			testRequest(secret1, "", true)
		})
		It("request should return the file path for a valid secret", func() {
			fakeStore.GetReturns(&state.Secret{Secret: secret1, Valid: true})
			expectedPath := path.Join(tmpSecretsDir, "test_secret1")

			testRequest(secret1, expectedPath, false)
		})

		It("request should return the file path for another valid secret", func() {
			fakeStore.GetReturns(&state.Secret{Secret: secret2, Valid: true})
			expectedPath := path.Join(tmpSecretsDir, "test_secret2")

			testRequest(secret2, expectedPath, false)
		})

		It("request should return an error and empty path when secret is invalid", func() {
			fakeStore.GetReturns(&state.Secret{Secret: invalidSecretType, Valid: false})

			testRequest(invalidSecretType, "", true)
		})

		It("should write all requested secrets", func() {
			err := memMgr.WriteAllRequestedSecrets()
			Expect(err).ToNot(HaveOccurred())

			expectedFileNames := []string{"test_secret1", "test_secret2"}

			// read all files from directory
			dir, err := ioutil.ReadDir(tmpSecretsDir)
			Expect(err).ToNot(HaveOccurred())

			// test that the files exist that we expect
			Expect(dir).To(HaveLen(2))
			actualFilenames := []string{dir[0].Name(), dir[1].Name()}
			Expect(actualFilenames).To(ConsistOf(expectedFileNames))
		})

		It("request should return the file path for secret after write", func() {
			fakeStore.GetReturns(&state.Secret{Secret: secret3, Valid: true})
			expectedPath := path.Join(tmpSecretsDir, "test_secret3")

			testRequest(secret3, expectedPath, false)
		})

		It("should write all requested secrets", func() {
			err := memMgr.WriteAllRequestedSecrets()
			Expect(err).ToNot(HaveOccurred())

			// read all files from directory
			dir, err := ioutil.ReadDir(tmpSecretsDir)
			Expect(err).ToNot(HaveOccurred())

			// only the secrets stored after the last write should be written to disk.
			Expect(dir).To(HaveLen(1))
			Expect(dir[0].Name()).To(Equal("test_secret3"))
		})
		When("no secrets are requested", func() {
			It("write all secrets should remove all existing secrets and write no additional secrets", func() {
				err := memMgr.WriteAllRequestedSecrets()
				Expect(err).ToNot(HaveOccurred())

				// read all files from directory
				dir, err := ioutil.ReadDir(tmpSecretsDir)
				Expect(err).ToNot(HaveOccurred())

				// no secrets should exist
				Expect(dir).To(BeEmpty())
			})
		})
	})
})

var _ = Describe("SecretStore", func() {
	var store state.SecretStore
	var invalidToValidSecret, validToInvalidSecret *apiv1.Secret

	BeforeEach(OncePerOrdered, func() {
		store = state.NewSecretStore()

		invalidToValidSecret = invalidSecretType.DeepCopy()
		invalidToValidSecret.Type = apiv1.SecretTypeTLS

		validToInvalidSecret = secret1.DeepCopy()
		validToInvalidSecret.Data[apiv1.TLSCertKey] = invalidCert

	})
	Describe("handles CRUD events on secrets", Ordered, func() {
		testUpsert := func(s *apiv1.Secret, valid bool) {
			store.Upsert(s)

			nsname := types.NamespacedName{Namespace: s.Namespace, Name: s.Name}
			actualSecret := store.Get(nsname)
			if valid {
				Expect(actualSecret.Valid).To(BeTrue())
			} else {
				Expect(actualSecret.Valid).To(BeFalse())
			}
			Expect(actualSecret.Secret).To(Equal(s))
		}

		testDelete := func(nsname types.NamespacedName) {
			store.Delete(nsname)

			s := store.Get(nsname)
			Expect(s).To(BeNil())
		}

		It("adds a new valid secret", func() {
			testUpsert(secret1, true)
		})
		It("adds another new valid secret", func() {
			testUpsert(secret2, true)
		})
		It("adds a secret with an invalid type", func() {
			testUpsert(invalidSecretType, false)
		})
		It("adds a secret with an invalid key", func() {
			testUpsert(invalidSecretKey, false)
		})
		It("deletes an invalid secret", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "invalid-key"}

			testDelete(nsname)
		})
		It("updates an invalid secret to valid", func() {
			testUpsert(invalidToValidSecret, true)
		})
		It("updates an valid secret to invalid (invalid cert)", func() {
			testUpsert(validToInvalidSecret, false)
		})
		It("deletes a secret", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "invalid-type"}

			testDelete(nsname)
		})
		It("deletes a secret", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "secret1"}

			testDelete(nsname)
		})
		It("gets remaining secret", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "secret2"}

			s := store.Get(nsname)
			Expect(s.Valid).To(BeTrue())
			Expect(s.Secret).To(Equal(secret2))
		})
		It("deletes final secret", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "secret2"}

			testDelete(nsname)
		})
		It("does not panic when secret is deleted that does not exist", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "dne"}

			store.Delete(nsname)
		})
	})
	Describe("File Management Error Cases", func() {
		var (
			fakeFileManager   *statefakes.FakeFileManager
			fakeStore         *statefakes.FakeSecretStore
			fakeFileInfoSlice []fs.FileInfo
			memMgr            state.SecretDiskMemoryManager
		)

		BeforeEach(OncePerOrdered, func() {
			fakeFileManager = &statefakes.FakeFileManager{}
			fakeStore = &statefakes.FakeSecretStore{}
			fakeFileInfoSlice = []fs.FileInfo{&statefakes.FakeFileInfo{}}
			memMgr = state.NewSecretDiskMemoryManager("", fakeStore, state.WithSecretFileManager(fakeFileManager))

			// populate a requested secret
			fakeStore.GetReturns(&state.Secret{Secret: secret1, Valid: true})
			_, err := memMgr.Request(types.NamespacedName{Namespace: secret1.Namespace, Name: secret1.Name})
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("Write all requested secrets", Ordered, func() {
			It("returns an error when secret directory cannot be read from", func() {
				errString := "read dir error"
				fakeFileManager.ReadDirReturns(nil, errors.New(errString))

				err := memMgr.WriteAllRequestedSecrets()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errString))
			})
			It("returns an error when a file cannot be removed from the secrets directory", func() {
				errString := "remove error"
				fakeFileManager.ReadDirReturns(fakeFileInfoSlice, nil)
				fakeFileManager.RemoveReturns(errors.New(errString))

				err := memMgr.WriteAllRequestedSecrets()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errString))
			})
			It("returns an error when file cannot be created", func() {
				errString := "create error"
				fakeFileManager.RemoveReturns(nil)
				fakeFileManager.CreateReturns(nil, errors.New(errString))

				err := memMgr.WriteAllRequestedSecrets()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errString))
			})
			It("returns an error when it cannot change the mode of the file", func() {
				errStr := "chmod error"
				fakeFileManager.CreateReturns(&os.File{}, nil)
				fakeFileManager.ChmodReturns(errors.New(errStr))

				err := memMgr.WriteAllRequestedSecrets()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errStr))
			})
			It("returns an error when file cannot be written to", func() {
				errString := "write error"
				fakeFileManager.ChmodReturns(nil)
				fakeFileManager.WriteReturns(errors.New(errString))

				err := memMgr.WriteAllRequestedSecrets()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errString))
			})
		})
	})
})
