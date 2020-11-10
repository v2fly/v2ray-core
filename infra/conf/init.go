package conf

func init() {
	RegisterConfigureFilePostProcessingStage("FakeDns", &FakeDnsPostProcessingStage{})
}
