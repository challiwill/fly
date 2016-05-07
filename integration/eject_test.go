package integration_test

import (
	"net/http"
	"os/exec"

	"github.com/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("EjectResource", func() {
	var (
		flyCmd *exec.Cmd
	)

	Context("when ATC request succeeds", func() {
		BeforeEach(func() {
			expectedGetURL := "/api/v1/pipelines/mypipeline/resources/myresource/versions"
			expectedDeleteURL := "/api/v1/pipelines/mypipeline/resources/myresource/versions/2/delete"
			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", expectedGetURL),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []atc.VersionedResource{
						atc.VersionedResource{
							ID:         1,
							PipelineID: 1,
							Type:       "some-type",
							Metadata:   nil,
							Resource:   "myresource",
							Version:    atc.Version{"other": "version"},
							Enabled:    true,
						},
						atc.VersionedResource{
							ID:         2,
							PipelineID: 2,
							Type:       "some-type",
							Metadata:   nil,
							Resource:   "myresource",
							Version:    atc.Version{"some": "version"},
							Enabled:    true,
						},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", expectedDeleteURL),
					ghttp.RespondWithJSONEncoded(http.StatusOK, ""),
				),
			)
		})

		It("sends delete resource request to ATC", func() {
			Expect(func() {
				flyCmd = exec.Command(flyPath, "-t", targetName, "eject-resource", "-r", "mypipeline/myresource", "-n", "some:version")
				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess).Should(gexec.Exit(0))

				Expect(sess.Out).To(gbytes.Say("deleted 'myresource' version 'some:version'"))

			})
		})
	})

	Context("when pipeline, resource, or version is not found", func() {
		BeforeEach(func() {
			expectedGetURL := "/api/v1/pipelines/mypipeline/resources/myresource/versions"
			expectedDeleteURL := "/api/v1/pipelines/mypipeline/resources/myresource/versions/2/delete"
			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", expectedGetURL),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []atc.VersionedResource{
						atc.VersionedResource{
							ID:         1,
							PipelineID: 1,
							Type:       "some-type",
							Metadata:   nil,
							Resource:   "myresource",
							Version:    atc.Version{"other": "version"},
							Enabled:    true,
						},
						atc.VersionedResource{
							ID:         2,
							PipelineID: 2,
							Type:       "some-type",
							Metadata:   nil,
							Resource:   "myresource",
							Version:    atc.Version{"some": "version"},
							Enabled:    true,
						},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", expectedDeleteURL),
					ghttp.RespondWithJSONEncoded(http.StatusNotFound, ""),
				),
			)
		})

		It("fails with error", func() {
			flyCmd = exec.Command(flyPath, "-t", targetName, "eject-resource", "-r", "mypipeline/myresource", "-n", "some:version")
			sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(1))

			Expect(sess.Err).To(gbytes.Say("pipeline 'mypipeline' or resource 'myresource' or version 'some:version' not found"))
		})
	})
})
