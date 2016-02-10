package elementalconductor

import (
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/NYTimes/encoding-wrapper/elementalconductor"
	"github.com/nytm/video-transcoding-api/config"
	"github.com/nytm/video-transcoding-api/provider"
)

func TestFactoryIsRegistered(t *testing.T) {
	_, err := provider.GetProviderFactory(Name)
	if err != nil {
		t.Fatal(err)
	}
}

func TestElementalConductorFactory(t *testing.T) {
	cfg := config.Config{
		ElementalConductor: &config.ElementalConductor{
			Host:        "elemental-server",
			UserLogin:   "myuser",
			APIKey:      "secret-key",
			AuthExpires: 30,
		},
	}
	provider, err := elementalConductorFactory(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	econductorProvider, ok := provider.(*elementalConductorProvider)
	if !ok {
		t.Fatalf("Wrong provider returned. Want elementalConductorProvider instance. Got %#v.", provider)
	}
	expected := elementalconductor.Client{
		Host:        "elemental-server",
		UserLogin:   "myuser",
		APIKey:      "secret-key",
		AuthExpires: 30,
	}
	if !reflect.DeepEqual(*econductorProvider.client, expected) {
		t.Errorf("Factory: wrong client returned. Want %#v. Got %#v.", expected, *econductorProvider.client)
	}
	if !reflect.DeepEqual(*econductorProvider.config, cfg) {
		t.Errorf("Factory: wrong config returned. Want %#v. Got %#v.", cfg, *econductorProvider.config)
	}
}

func TestElementalConductorFactoryValidation(t *testing.T) {
	var tests = []struct {
		host        string
		userLogin   string
		apiKey      string
		authExpires int
	}{
		{"", "", "", 0},
		{"myhost", "", "", 0},
		{"", "myuser", "", 0},
		{"", "", "mykey", 0},
		{"", "", "", 30},
	}
	for _, test := range tests {
		cfg := config.Config{
			ElementalConductor: &config.ElementalConductor{
				Host:        test.host,
				UserLogin:   test.userLogin,
				APIKey:      test.apiKey,
				AuthExpires: test.authExpires,
			},
		}
		provider, err := elementalConductorFactory(&cfg)
		if provider != nil {
			t.Errorf("Unexpected non-nil provider: %#v", provider)
		}
		if err != errElementalConductorInvalidConfig {
			t.Errorf("Wrong error returned. Want errElementalConductorInvalidConfig. Got %#v", err)
		}
	}
}

func TestElementalNewJob(t *testing.T) {
	elementalConductorConfig := config.Config{
		ElementalConductor: &config.ElementalConductor{
			Host:            "https://mybucket.s3.amazonaws.com/destination-dir/",
			UserLogin:       "myuser",
			APIKey:          "elemental-api-key",
			AuthExpires:     30,
			AccessKeyID:     "aws-access-key",
			SecretAccessKey: "aws-secret-key",
			Destination:     "s3://destination",
		},
	}
	prov, err := elementalConductorFactory(&elementalConductorConfig)
	if err != nil {
		t.Fatal(err)
	}
	presetProvider, ok := prov.(*elementalConductorProvider)
	if !ok {
		t.Fatal("Could not type assert test provider to elementalConductorProvider")
	}
	source := "http://some.nice/video.mov"
	presets := []string{"15", "20"}
	newJob := presetProvider.newJob(source, presets)

	expectedJob := elementalconductor.Job{
		XMLName: xml.Name{
			Local: "job",
		},
		Input: elementalconductor.Input{
			FileInput: elementalconductor.Location{
				URI:      "http://some.nice/video.mov",
				Username: "aws-access-key",
				Password: "aws-secret-key",
			},
		},
		Priority: 50,
		OutputGroup: elementalconductor.OutputGroup{
			Order: 1,
			FileGroupSettings: elementalconductor.FileGroupSettings{
				Destination: elementalconductor.Location{
					URI:      "s3://destination/video",
					Username: "aws-access-key",
					Password: "aws-secret-key",
				},
			},
			Type: "file_group_settings",
			Output: []elementalconductor.Output{
				{
					StreamAssemblyName: "stream_0",
					NameModifier:       "_15",
					Order:              0,
					Extension:          ".mp4",
				},
				{
					StreamAssemblyName: "stream_1",
					NameModifier:       "_20",
					Order:              1,
					Extension:          ".mp4",
				},
			},
		},
		StreamAssembly: []elementalconductor.StreamAssembly{
			{
				Name:   "stream_0",
				Preset: "15",
			},
			{
				Name:   "stream_1",
				Preset: "20",
			},
		},
	}
	if !reflect.DeepEqual(&expectedJob, newJob) {
		t.Errorf("New job not according to spec.\nWanted %v.\nGot    %v.", &expectedJob, newJob)
	}
}

func TestJobStatusMap(t *testing.T) {
	var tests = []struct {
		elementalConductorStatus string
		expected                 provider.Status
	}{
		{"pending", provider.StatusQueued},
		{"preprocessing", provider.StatusStarted},
		{"running", provider.StatusStarted},
		{"postprocessing", provider.StatusStarted},
		{"complete", provider.StatusFinished},
		{"cancelled", provider.StatusCanceled},
		{"error", provider.StatusFailed},
		{"unknown", provider.StatusUnknown},
		{"someotherstatus", provider.StatusUnknown},
	}
	var p elementalConductorProvider
	for _, test := range tests {
		got := p.statusMap(test.elementalConductorStatus)
		if got != test.expected {
			t.Errorf("statusMap(%q): wrong value. Want %q. Got %q", test.elementalConductorStatus, test.expected, got)
		}
	}
}
