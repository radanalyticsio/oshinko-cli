package util

import (
	"fmt"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

//CorruptImage is a helper that tags the image to be corrupted, the corruptee, as the corruptor string, resulting in the wrong image being used when corruptee is referenced later on;  strategy is for ginkgo debug; ginkgo error checking leveraged
func CorruptImage(corruptee, corruptor string) {
	fmt.Fprintf(g.GinkgoWriter, "Calling docker tag to corrupt builder image %s by tagging %s", corruptee, corruptor)

	cerr := TagImage(corruptee, corruptor)

	fmt.Fprintf(g.GinkgoWriter, "Tagging %s to %s complete with err %v", corruptor, corruptee, cerr)
	o.Expect(cerr).NotTo(o.HaveOccurred())
	VerifyImagesSame(corruptee, corruptor, "image corruption")
}

//ResetImage is a helper the allows the programmer to undo any corruption performed by CorruptImage; ginkgo error checking leveraged
func ResetImage(tags map[string]string) {
	fmt.Fprintf(g.GinkgoWriter, "Calling docker tag to reset images")

	for corruptedTag, goodTag := range tags {
		err := TagImage(corruptedTag, goodTag)
		fmt.Fprintf(g.GinkgoWriter, "Reset for %s to %s complete with err %v", corruptedTag, goodTag, err)
		o.Expect(err).NotTo(o.HaveOccurred())
	}

}

//DumpImage is a helper that inspects the image along with some ginkgo debug
func DumpImage(name string) {
	fmt.Fprintf(g.GinkgoWriter, "Calling docker inspect for image %s", name)

	image, err := InspectImage(name)
	o.Expect(err).NotTo(o.HaveOccurred())
	if image != nil {
		fmt.Fprintf(g.GinkgoWriter, "Returned docker image %+v", image)
		fmt.Fprintf(g.GinkgoWriter, "Container config %+v and user %s", image.ContainerConfig, image.ContainerConfig.User)
		if image.Config != nil {
			fmt.Fprintf(g.GinkgoWriter, "Image config %+v and user %s", image.Config, image.Config.User)
		}
	}
}

//DumpAndReturnTagging takes and array of tags and obtains the hex image IDs, dumps them to ginkgo for printing, and then returns them
func DumpAndReturnTagging(tags []string) ([]string, error) {
	hexIDs, err := GetImageIDForTags(tags)
	if err != nil {
		return nil, err
	}
	for i, hexID := range hexIDs {
		fmt.Fprintf(g.GinkgoWriter, "tag %s hex id %s ", tags[i], hexID)
	}
	return hexIDs, nil
}

//VerifyImagesSame will take the two supplied image tags and see if they reference the same hexadecimal image ID;  strategy is for debug
func VerifyImagesSame(comp1, comp2, strategy string) {
	tag1 := comp1 + ":latest"
	tag2 := comp2 + ":latest"

	comps := []string{tag1, tag2}
	retIDs, gerr := GetImageIDForTags(comps)

	o.Expect(gerr).NotTo(o.HaveOccurred())
	fmt.Fprintf(g.GinkgoWriter, "%s  compare image - %s, %s, %s, %s", strategy, tag1, tag2, retIDs[0], retIDs[1])
	o.Ω(len(retIDs[0])).Should(o.BeNumerically(">", 0))
	o.Ω(len(retIDs[1])).Should(o.BeNumerically(">", 0))
	o.Ω(retIDs[0]).Should(o.Equal(retIDs[1]))
}

//VerifyImagesDifferent will that the two supplied image tags and see if they reference different hexadecimal image IDs; strategy is for ginkgo debug, also leverage ginkgo error checking
func VerifyImagesDifferent(comp1, comp2, strategy string) {
	tag1 := comp1 + ":latest"
	tag2 := comp2 + ":latest"

	comps := []string{tag1, tag2}
	retIDs, gerr := GetImageIDForTags(comps)

	o.Expect(gerr).NotTo(o.HaveOccurred())
	fmt.Fprintf(g.GinkgoWriter, "%s  compare image - %s, %s, %s, %s", strategy, tag1, tag2, retIDs[0], retIDs[1])
	o.Ω(len(retIDs[0])).Should(o.BeNumerically(">", 0))
	o.Ω(len(retIDs[1])).Should(o.BeNumerically(">", 0))
	o.Ω(retIDs[0] != retIDs[1]).Should(o.BeTrue())
}

//CreateResource creates the resources from the supplied json file (not a template); ginkgo error checking included
func CreateResource(jsonFilePath string, oc *CLI) error {
	err := oc.Run("create").Args("-f", jsonFilePath).Execute()
	return err
}
