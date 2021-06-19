package v4

func init() {
	RegisterConfigureFilePostProcessingStage("FakeDNS", &FakeDNSPostProcessingStage{})
}
