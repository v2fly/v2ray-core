package conf

func init() {
	RegisterConfigureFilePostProcessingStage("FakeDns", &FakeDNSPostProcessingStage{})
}
