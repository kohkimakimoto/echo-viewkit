package pongo2

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponentHTMLTagPreProcessor(t *testing.T) {
	var defaultConfig = ComponentHTMLTagPreProcessorConfig{
		TagPrefix: "x-",
	}

	testCases := []struct {
		config ComponentHTMLTagPreProcessorConfig
		input  string
		output string
	}{
		{
			config: defaultConfig,
			input:  `<x-alert></x-alert>`,
			output: `{% component "alert" %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-ui.alert></x-ui.alert>`,
			output: `{% component "ui.alert" %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert key="value" :key2="value2"></x-alert>`,
			output: `{% component "alert" withAttrs "key"="value" "key2"=value2 %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert slot-data="{aaa,bbb}" key="value" :key2="value2"></x-alert>`,
			output: `{% component "alert" slotData="{aaa,bbb}" withAttrs "key"="value" "key2"=value2 %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert key-aaa="value" :key2-bbb="value2" key3Ccc="value3">Inner content</x-alert>`,
			output: `{% component "alert" withAttrs "key-aaa"="value" "key2-bbb"=value2 "key3Ccc"="value3" %}Inner content{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert key="value" :key2="value2" />`,
			output: `{% component "alert" withAttrs "key"="value" "key2"=value2 %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-ui.alert key="value" :key2="value2" />`,
			output: `{% component "ui.alert" withAttrs "key"="value" "key2"=value2 %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert slot-data="{aa,bb}" key="value" :key2="value2" />`,
			output: `{% component "alert" slotData="{aa,bb}" withAttrs "key"="value" "key2"=value2 %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input: `<x-alert
                      slot-data="{aa,bb}"
                      key-aaa="value"
                      :key2-bbb="value2"
                    >
                      Inner content
                    </x-alert>`,
			output: `{% component "alert" slotData="{aa,bb}" withAttrs "key-aaa"="value" "key2-bbb"=value2 %}
                      Inner content
                    {% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input: `<x-alert>
                      <x-slot name="title">Title</x-slot>
					</x-alert>`,
			output: `{% component "alert" %}
                      {% slot "title" %}Title{% endslot %}
					{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input: `<x-alert1>
                      <x-alert2>
                        <x-alert3>text</x-alert3>
                      </x-alert2>
                    </x-alert>`,
			output: `{% component "alert1" %}
                      {% component "alert2" %}
                        {% component "alert3" %}text{% endcomponent %}
                      {% endcomponent %}
                    {% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert><img /></x-alert>`,
			output: `{% component "alert" %}<img />{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert hoge="aa">aaa</x-alert>`,
			output: `{% component "alert" withAttrs "hoge"="aa" %}aaa{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input: `<x-alert
                      hoge="aa"
                    />`,
			output: `{% component "alert" withAttrs "hoge"="aa" %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert hoge="aa"><x-slot name="aaa">hoge</x-slot></x-alert>`,
			output: `{% component "alert" withAttrs "hoge"="aa" %}{% slot "aaa" %}hoge{% endslot %}{% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input: `<x-alert hoge="aa">
                      <x-slot name="aaa">hoge1</x-slot>
                      <x-slot name="bbb">hoge2</x-slot>
                    </x-alert>`,
			output: `{% component "alert" withAttrs "hoge"="aa" %}
                      {% slot "aaa" %}hoge1{% endslot %}
                      {% slot "bbb" %}hoge2{% endslot %}
                    {% endcomponent %}`,
		},
		{
			config: defaultConfig,
			input:  `{% verbatim %}<x-alert>aaa</x-alert>{% endverbatim %}`,
			output: `{% verbatim %}<x-alert>aaa</x-alert>{% endverbatim %}`,
		},
		{
			config: defaultConfig,
			input: `
{% verbatim %}
<x-alert />
<x-alert>
hogehoge
</x-alert>
{% endverbatim %}`,
			output: `
{% verbatim %}
<x-alert />
<x-alert>
hogehoge
</x-alert>
{% endverbatim %}`,
		},
		{
			config: defaultConfig,
			input:  `<x-alert messageFoo="aaa" message-bar="bbb"></x-alert>`,
			output: `{% component "alert" withAttrs "messageFoo"="aaa" "message-bar"="bbb" %}{% endcomponent %}`,
		},
	}

	for _, tc := range testCases {
		p := ComponentHTMLTagPreProcessor(tc.config)
		src := bytes.NewBufferString(tc.input)
		dst := new(bytes.Buffer)
		err := p.Execute(dst, src)
		assert.NoError(t, err)
		assert.Equal(t, tc.output, dst.String())
	}
}
