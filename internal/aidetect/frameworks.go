// Package aidetect discovers AI/ML evaluation frameworks, prompt libraries,
// and dataset infrastructure in a repository. It auto-derives scenarios from
// code so users never have to write YAML scenario declarations manually.
//
// Supported frameworks:
//   - promptfoo (config-based prompt evaluation)
//   - deepeval (Python eval framework)
//   - ragas (RAG evaluation)
//   - langchain / langsmith (LLM application framework + observability)
//   - openai (OpenAI SDK patterns)
//   - anthropic (Anthropic SDK patterns)
//   - llamaindex (LLM data framework)
//   - huggingface (transformers/datasets)
//   - pytest-eval (pytest files in eval directories)
//   - vitest-eval (vitest files in eval directories)
package aidetect

// Framework identifies a detected AI/ML evaluation framework.
type Framework struct {
	// Name is the canonical framework identifier.
	Name string `json:"name"`

	// Version is the detected version constraint, if available.
	Version string `json:"version,omitempty"`

	// Source describes how the framework was detected.
	Source string `json:"source"` // "config", "dependency", "import", "directory"

	// ConfigFile is the path to the framework's config file, if any.
	ConfigFile string `json:"configFile,omitempty"`

	// Confidence is how confident we are in the detection (0.0-1.0).
	Confidence float64 `json:"confidence"`
}

// FrameworkSignature defines how to detect a specific framework.
type FrameworkSignature struct {
	Name        string
	ConfigFiles []string   // Config file names to look for
	DependencyKeys []string // package.json or pyproject.toml dependency names
	ImportPatterns []string // Import patterns in source code
	DirectoryPatterns []string // Directory names that indicate this framework
}

// KnownFrameworks lists all supported AI/ML framework signatures.
var KnownFrameworks = []FrameworkSignature{
	{
		Name:        "promptfoo",
		ConfigFiles: []string{"promptfooconfig.yaml", "promptfooconfig.yml", "promptfoo.yaml", ".promptfoo.yaml"},
		DependencyKeys: []string{"promptfoo"},
		ImportPatterns: []string{"promptfoo", "from promptfoo"},
	},
	{
		Name:        "deepeval",
		DependencyKeys: []string{"deepeval"},
		ImportPatterns: []string{"from deepeval", "import deepeval"},
	},
	{
		Name:        "ragas",
		DependencyKeys: []string{"ragas"},
		ImportPatterns: []string{"from ragas", "import ragas"},
	},
	{
		Name:        "langchain",
		DependencyKeys: []string{"langchain", "@langchain/core", "@langchain/openai", "@langchain/anthropic", "langchain-core", "langchain-openai", "langchain-anthropic"},
		ImportPatterns: []string{"from langchain", "import langchain", "@langchain/", "langchain/"},
	},
	{
		Name:        "langsmith",
		DependencyKeys: []string{"langsmith", "@langchain/smith"},
		ImportPatterns: []string{"from langsmith", "import langsmith", "LANGSMITH_API_KEY", "@langchain/smith"},
		ConfigFiles: []string{".langsmith.yaml", "langsmith.config.ts", "langsmith.config.js"},
	},
	{
		Name:        "openai",
		DependencyKeys: []string{"openai", "@openai/api"},
		ImportPatterns: []string{"from openai", "import openai", "import OpenAI", "from openai import"},
	},
	{
		Name:        "anthropic",
		DependencyKeys: []string{"@anthropic-ai/sdk", "anthropic"},
		ImportPatterns: []string{"from anthropic", "import anthropic", "import Anthropic", "@anthropic-ai/sdk"},
	},
	{
		Name:        "llamaindex",
		DependencyKeys: []string{"llamaindex", "llama-index", "llama_index"},
		ImportPatterns: []string{"from llama_index", "from llamaindex", "import llamaindex"},
	},
	{
		Name:        "huggingface",
		DependencyKeys: []string{"transformers", "datasets", "@huggingface/inference"},
		ImportPatterns: []string{"from transformers", "from datasets", "import transformers", "@huggingface/"},
	},
	{
		Name:        "vertexai",
		DependencyKeys: []string{"@google-cloud/vertexai", "google-cloud-aiplatform"},
		ImportPatterns: []string{"from google.cloud import aiplatform", "from vertexai", "@google-cloud/vertexai"},
	},
	{
		Name:        "aws-bedrock",
		DependencyKeys: []string{"@aws-sdk/client-bedrock-runtime", "boto3"},
		ImportPatterns: []string{"bedrock-runtime", "from boto3", "invoke_model"},
	},
}
