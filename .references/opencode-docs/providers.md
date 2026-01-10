---
title: "Providers | OpenCode"
source: "https://opencode.ai/docs/providers/"
crawled: "2026-01-05T00:46:23Z"
---

[Skip to content](https://opencode.ai/docs/providers/#_top)

# Providers

Using any LLM provider in OpenCode.

OpenCode uses the [AI SDK](https://ai-sdk.dev/) and [Models.dev](https://models.dev/) to support for **75+ LLM providers** and it supports running local models.

To add a provider you need to:

1. Add the API keys for the provider using the `/connect` command.
2. Configure the provider in your OpenCode config.

* * *

### [Credentials](https://opencode.ai/docs/providers/\#credentials)

When you add a provider’s API keys with the `/connect` command, they are stored
in `~/.local/share/opencode/auth.json`.

* * *

### [Config](https://opencode.ai/docs/providers/\#config)

You can customize the providers through the `provider` section in your OpenCode
config.

* * *

#### [Base URL](https://opencode.ai/docs/providers/\#base-url)

You can customize the base URL for any provider by setting the `baseURL` option. This is useful when using proxy services or custom endpoints.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "anthropic": {

      "options": {

        "baseURL": "https://api.anthropic.com/v1"

      }

    }

  }

}
```

* * *

## [OpenCode Zen](https://opencode.ai/docs/providers/\#opencode-zen)

OpenCode Zen is a list of models provided by the OpenCode team that have been
tested and verified to work well with OpenCode. [Learn more](https://opencode.ai/docs/zen).

1. Run the `/connect` command in the TUI, select opencode, and head to [opencode.ai/auth](https://opencode.ai/auth).



```
/connect
```

2. Sign in, add your billing details, and copy your API key.

3. Paste your API key.



```
┌ API key

│

│

└ enter
```

4. Run `/models` in the TUI to see the list of models we recommend.



```
/models
```


It works like any other provider in OpenCode. And is completely optional to use
it.

* * *

## [Directory](https://opencode.ai/docs/providers/\#directory)

Let’s look at some of the providers in detail. If you’d like to add a provider to the
list, feel free to open a PR.

* * *

### [Amazon Bedrock](https://opencode.ai/docs/providers/\#amazon-bedrock)

To use Amazon Bedrock with OpenCode:

1. Head over to the **Model catalog** in the Amazon Bedrock console and request
access to the models you want.

2. You’ll need either to set one of the following environment variables:


   - `AWS_ACCESS_KEY_ID`: You can get this by creating an IAM user and generating
     an access key for it.
   - `AWS_PROFILE`: First login through AWS IAM Identity Center (or AWS SSO) using
     `aws sso login`. Then get the name of the profile you want to use.
   - `AWS_BEARER_TOKEN_BEDROCK`: You can generate a long-term API key from the
     Amazon Bedrock console.

Once you have one of the above, set it while running opencode.

```
AWS_ACCESS_KEY_ID=XXX opencode
```

Or add it to your bash profile.

```
export AWS_ACCESS_KEY_ID=XXX
```

3. Run the `/models` command to select the model you want.



```
/models
```


* * *

### [Anthropic](https://opencode.ai/docs/providers/\#anthropic)

We recommend signing up for [Claude Pro](https://www.anthropic.com/news/claude-pro) or [Max](https://www.anthropic.com/max).

1. Once you’ve signed up, run the `/connect` command and select Anthropic.



```
/connect
```

2. Here you can select the **Claude Pro/Max** option and it’ll open your browser
and ask you to authenticate.



```
┌ Select auth method

│

│ Claude Pro/Max

│ Create an API Key

│ Manually enter API Key

└
```

3. Now all the the Anthropic models should be available when you use the `/models` command.



```
/models
```


##### [Using API keys](https://opencode.ai/docs/providers/\#using-api-keys)

You can also select **Create an API Key** if you don’t have a Pro/Max subscription. It’ll also open your browser and ask you to login to Anthropic and give you a code you can paste in your terminal.

Or if you already have an API key, you can select **Manually enter API Key** and paste it in your terminal.

* * *

### [Azure OpenAI](https://opencode.ai/docs/providers/\#azure-openai)

1. Head over to the [Azure portal](https://portal.azure.com/) and create an **Azure OpenAI** resource. You’ll need:
   - **Resource name**: This becomes part of your API endpoint (`https://RESOURCE_NAME.openai.azure.com/`)
   - **API key**: Either `KEY 1` or `KEY 2` from your resource
2. Go to [Azure AI Foundry](https://ai.azure.com/) and deploy a model.

3. Run the `/connect` command and search for **Azure**.



```
/connect
```

4. Enter your API key.



```
┌ API key

│

│

└ enter
```

5. Set your resource name as an environment variable:



```
AZURE_RESOURCE_NAME=XXX opencode
```









Or add it to your bash profile:



```
export AZURE_RESOURCE_NAME=XXX
```

6. Run the `/models` command to select your deployed model.



```
/models
```


* * *

### [Azure Cognitive Services](https://opencode.ai/docs/providers/\#azure-cognitive-services)

1. Head over to the [Azure portal](https://portal.azure.com/) and create an **Azure OpenAI** resource. You’ll need:
   - **Resource name**: This becomes part of your API endpoint (`https://AZURE_COGNITIVE_SERVICES_RESOURCE_NAME.cognitiveservices.azure.com/`)
   - **API key**: Either `KEY 1` or `KEY 2` from your resource
2. Go to [Azure AI Foundry](https://ai.azure.com/) and deploy a model.

3. Run the `/connect` command and search for **Azure Cognitive Services**.



```
/connect
```

4. Enter your API key.



```
┌ API key

│

│

└ enter
```

5. Set your resource name as an environment variable:



```
AZURE_COGNITIVE_SERVICES_RESOURCE_NAME=XXX opencode
```









Or add it to your bash profile:



```
export AZURE_COGNITIVE_SERVICES_RESOURCE_NAME=XXX
```

6. Run the `/models` command to select your deployed model.



```
/models
```


* * *

### [Baseten](https://opencode.ai/docs/providers/\#baseten)

1. Head over to the [Baseten](https://app.baseten.co/), create an account, and generate an API key.

2. Run the `/connect` command and search for **Baseten**.



```
/connect
```

3. Enter your Baseten API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model.



```
/models
```


* * *

### [Cerebras](https://opencode.ai/docs/providers/\#cerebras)

1. Head over to the [Cerebras console](https://inference.cerebras.ai/), create an account, and generate an API key.

2. Run the `/connect` command and search for **Cerebras**.



```
/connect
```

3. Enter your Cerebras API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Qwen 3 Coder 480B_.



```
/models
```


* * *

### [Cloudflare AI Gateway](https://opencode.ai/docs/providers/\#cloudflare-ai-gateway)

Cloudflare AI Gateway lets you access models from OpenAI, Anthropic, Workers AI, and more through a unified endpoint. With [Unified Billing](https://developers.cloudflare.com/ai-gateway/features/unified-billing/) you don’t need separate API keys for each provider.

1. Head over to the [Cloudflare dashboard](https://dash.cloudflare.com/), navigate to **AI** \> **AI Gateway**, and create a new gateway.

2. Set your Account ID and Gateway ID as environment variables.



```
export CLOUDFLARE_ACCOUNT_ID=your-32-character-account-id

export CLOUDFLARE_GATEWAY_ID=your-gateway-id
```

3. Run the `/connect` command and search for **Cloudflare AI Gateway**.



```
/connect
```

4. Enter your Cloudflare API token.



```
┌ API key

│

│

└ enter
```









Or set it as an environment variable.



```
export CLOUDFLARE_API_TOKEN=your-api-token
```

5. Run the `/models` command to select a model.



```
/models
```









You can also add models through your opencode config.



```
{

     "$schema": "https://opencode.ai/config.json",

     "provider": {

       "cloudflare-ai-gateway": {

         "models": {

           "openai/gpt-4o": {},

           "anthropic/claude-sonnet-4": {}

         }

       }

     }

}
```


* * *

### [Cortecs](https://opencode.ai/docs/providers/\#cortecs)

1. Head over to the [Cortecs console](https://cortecs.ai/), create an account, and generate an API key.

2. Run the `/connect` command and search for **Cortecs**.



```
/connect
```

3. Enter your Cortecs API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Kimi K2 Instruct_.



```
/models
```


* * *

### [DeepSeek](https://opencode.ai/docs/providers/\#deepseek)

1. Head over to the [DeepSeek console](https://platform.deepseek.com/), create an account, and click **Create new API key**.

2. Run the `/connect` command and search for **DeepSeek**.



```
/connect
```

3. Enter your DeepSeek API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a DeepSeek model like _DeepSeek Reasoner_.



```
/models
```


* * *

### [Deep Infra](https://opencode.ai/docs/providers/\#deep-infra)

1. Head over to the [Deep Infra dashboard](https://deepinfra.com/dash), create an account, and generate an API key.

2. Run the `/connect` command and search for **Deep Infra**.



```
/connect
```

3. Enter your Deep Infra API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model.



```
/models
```


* * *

### [Fireworks AI](https://opencode.ai/docs/providers/\#fireworks-ai)

1. Head over to the [Fireworks AI console](https://app.fireworks.ai/), create an account, and click **Create API Key**.

2. Run the `/connect` command and search for **Fireworks AI**.



```
/connect
```

3. Enter your Fireworks AI API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Kimi K2 Instruct_.



```
/models
```


* * *

### [GitHub Copilot](https://opencode.ai/docs/providers/\#github-copilot)

To use your GitHub Copilot subscription with opencode:

1. Run the `/connect` command and search for GitHub Copilot.



```
/connect
```

2. Navigate to [github.com/login/device](https://github.com/login/device) and enter the code.



```
┌ Login with GitHub Copilot

│

│ https://github.com/login/device

│

│ Enter code: 8F43-6FCF

│

└ Waiting for authorization...
```

3. Now run the `/models` command to select the model you want.



```
/models
```


* * *

### [Google Vertex AI](https://opencode.ai/docs/providers/\#google-vertex-ai)

To use Google Vertex AI with OpenCode:

1. Head over to the **Model Garden** in the Google Cloud Console and check the
models available in your region.

2. Set the required environment variables:


   - `GOOGLE_CLOUD_PROJECT`: Your Google Cloud project ID
   - `VERTEX_LOCATION` (optional): The region for Vertex AI (defaults to `global`)
   - Authentication (choose one):
     - `GOOGLE_APPLICATION_CREDENTIALS`: Path to your service account JSON key file
     - Authenticate using gcloud CLI: `gcloud auth application-default login`

Set them while running opencode.

```
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json GOOGLE_CLOUD_PROJECT=your-project-id opencode
```

Or add them to your bash profile.

```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

export GOOGLE_CLOUD_PROJECT=your-project-id

export VERTEX_LOCATION=global
```

3. Run the `/models` command to select the model you want.



```
/models
```


* * *

### [Groq](https://opencode.ai/docs/providers/\#groq)

1. Head over to the [Groq console](https://console.groq.com/), click **Create API Key**, and copy the key.

2. Run the `/connect` command and search for Groq.



```
/connect
```

3. Enter the API key for the provider.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select the one you want.



```
/models
```


* * *

### [Hugging Face](https://opencode.ai/docs/providers/\#hugging-face)

[Hugging Face Inference Providers](https://huggingface.co/docs/inference-providers) provides access to open models supported by 17+ providers.

1. Head over to [Hugging Face settings](https://huggingface.co/settings/tokens/new?ownUserPermissions=inference.serverless.write&tokenType=fineGrained) to create a token with permission to make calls to Inference Providers.

2. Run the `/connect` command and search for **Hugging Face**.



```
/connect
```

3. Enter your Hugging Face token.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Kimi-K2-Instruct_ or _GLM-4.6_.



```
/models
```


* * *

### [Helicone](https://opencode.ai/docs/providers/\#helicone)

[Helicone](https://helicone.ai/) is an LLM observability platform that provides logging, monitoring, and analytics for your AI applications. The Helicone AI Gateway routes your requests to the appropriate provider automatically based on the model.

1. Head over to [Helicone](https://helicone.ai/), create an account, and generate an API key from your dashboard.

2. Run the `/connect` command and search for **Helicone**.



```
/connect
```

3. Enter your Helicone API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model.



```
/models
```


For more providers and advanced features like caching and rate limiting, check the [Helicone documentation](https://docs.helicone.ai/).

#### [Optional Configs](https://opencode.ai/docs/providers/\#optional-configs)

In the event you see a feature or model from Helicone that isn’t configured automatically through opencode, you can always configure it yourself.

Here’s [Helicone’s Model Directory](https://helicone.ai/models), you’ll need this to grab the IDs of the models you want to add.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "helicone": {

      "npm": "@ai-sdk/openai-compatible",

      "name": "Helicone",

      "options": {

        "baseURL": "https://ai-gateway.helicone.ai",

      },

      "models": {

        "gpt-4o": {

          // Model ID (from Helicone's model directory page)

          "name": "GPT-4o", // Your own custom name for the model

        },

        "claude-sonnet-4-20250514": {

          "name": "Claude Sonnet 4",

        },

      },

    },

  },

}
```

#### [Custom Headers](https://opencode.ai/docs/providers/\#custom-headers)

Helicone supports custom headers for features like caching, user tracking, and session management. Add them to your provider config using `options.headers`:

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "helicone": {

      "npm": "@ai-sdk/openai-compatible",

      "name": "Helicone",

      "options": {

        "baseURL": "https://ai-gateway.helicone.ai",

        "headers": {

          "Helicone-Cache-Enabled": "true",

          "Helicone-User-Id": "opencode",

        },

      },

    },

  },

}
```

##### [Session tracking](https://opencode.ai/docs/providers/\#session-tracking)

Helicone’s [Sessions](https://docs.helicone.ai/features/sessions) feature lets you group related LLM requests together. Use the [opencode-helicone-session](https://github.com/H2Shami/opencode-helicone-session) plugin to automatically log each OpenCode conversation as a session in Helicone.

```
npm install -g opencode-helicone-session
```

Add it to your config.

```
{

  "plugin": ["opencode-helicone-session"]

}
```

The plugin injects `Helicone-Session-Id` and `Helicone-Session-Name` headers into your requests. In Helicone’s Sessions page, you’ll see each OpenCode conversation listed as a separate session.

##### [Common Helicone headers](https://opencode.ai/docs/providers/\#common-helicone-headers)

| Header | Description |
| --- | --- |
| `Helicone-Cache-Enabled` | Enable response caching (`true`/`false`) |
| `Helicone-User-Id` | Track metrics by user |
| `Helicone-Property-[Name]` | Add custom properties (e.g., `Helicone-Property-Environment`) |
| `Helicone-Prompt-Id` | Associate requests with prompt versions |

See the [Helicone Header Directory](https://docs.helicone.ai/helicone-headers/header-directory) for all available headers.

* * *

### [llama.cpp](https://opencode.ai/docs/providers/\#llamacpp)

You can configure opencode to use local models through [llama.cpp’s](https://github.com/ggml-org/llama.cpp) llama-server utility

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "llama.cpp": {

      "npm": "@ai-sdk/openai-compatible",

      "name": "llama-server (local)",

      "options": {

        "baseURL": "http://127.0.0.1:8080/v1"

      },

      "models": {

        "qwen3-coder:a3b": {

          "name": "Qwen3-Coder: a3b-30b (local)",

          "limit": {

            "context": 128000,

            "output": 65536

          }

        }

      }

    }

  }

}
```

In this example:

- `llama.cpp` is the custom provider ID. This can be any string you want.
- `npm` specifies the package to use for this provider. Here, `@ai-sdk/openai-compatible` is used for any OpenAI-compatible API.
- `name` is the display name for the provider in the UI.
- `options.baseURL` is the endpoint for the local server.
- `models` is a map of model IDs to their configurations. The model name will be displayed in the model selection list.

* * *

### [IO.NET](https://opencode.ai/docs/providers/\#ionet)

IO.NET offers 17 models optimized for various use cases:

1. Head over to the [IO.NET console](https://ai.io.net/), create an account, and generate an API key.

2. Run the `/connect` command and search for **IO.NET**.



```
/connect
```

3. Enter your IO.NET API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model.



```
/models
```


* * *

### [LM Studio](https://opencode.ai/docs/providers/\#lm-studio)

You can configure opencode to use local models through LM Studio.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "lmstudio": {

      "npm": "@ai-sdk/openai-compatible",

      "name": "LM Studio (local)",

      "options": {

        "baseURL": "http://127.0.0.1:1234/v1"

      },

      "models": {

        "google/gemma-3n-e4b": {

          "name": "Gemma 3n-e4b (local)"

        }

      }

    }

  }

}
```

In this example:

- `lmstudio` is the custom provider ID. This can be any string you want.
- `npm` specifies the package to use for this provider. Here, `@ai-sdk/openai-compatible` is used for any OpenAI-compatible API.
- `name` is the display name for the provider in the UI.
- `options.baseURL` is the endpoint for the local server.
- `models` is a map of model IDs to their configurations. The model name will be displayed in the model selection list.

* * *

### [Moonshot AI](https://opencode.ai/docs/providers/\#moonshot-ai)

To use Kimi K2 from Moonshot AI:

1. Head over to the [Moonshot AI console](https://platform.moonshot.ai/console), create an account, and click **Create API key**.

2. Run the `/connect` command and search for **Moonshot AI**.



```
/connect
```

3. Enter your Moonshot API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select _Kimi K2_.



```
/models
```


* * *

### [MiniMax](https://opencode.ai/docs/providers/\#minimax)

1. Head over to the [MiniMax API Console](https://platform.minimax.io/login), create an account, and generate an API key.

2. Run the `/connect` command and search for **MiniMax**.



```
/connect
```

3. Enter your MiniMax API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _M2.1_.



```
/models
```


* * *

### [Nebius Token Factory](https://opencode.ai/docs/providers/\#nebius-token-factory)

1. Head over to the [Nebius Token Factory console](https://tokenfactory.nebius.com/), create an account, and click **Add Key**.

2. Run the `/connect` command and search for **Nebius Token Factory**.



```
/connect
```

3. Enter your Nebius Token Factory API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Kimi K2 Instruct_.



```
/models
```


* * *

### [Ollama](https://opencode.ai/docs/providers/\#ollama)

You can configure opencode to use local models through Ollama.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "ollama": {

      "npm": "@ai-sdk/openai-compatible",

      "name": "Ollama (local)",

      "options": {

        "baseURL": "http://localhost:11434/v1"

      },

      "models": {

        "llama2": {

          "name": "Llama 2"

        }

      }

    }

  }

}
```

In this example:

- `ollama` is the custom provider ID. This can be any string you want.
- `npm` specifies the package to use for this provider. Here, `@ai-sdk/openai-compatible` is used for any OpenAI-compatible API.
- `name` is the display name for the provider in the UI.
- `options.baseURL` is the endpoint for the local server.
- `models` is a map of model IDs to their configurations. The model name will be displayed in the model selection list.

* * *

### [Ollama Cloud](https://opencode.ai/docs/providers/\#ollama-cloud)

To use Ollama Cloud with OpenCode:

1. Head over to [https://ollama.com/](https://ollama.com/) and sign in or create an account.

2. Navigate to **Settings** \> **Keys** and click **Add API Key** to generate a new API key.

3. Copy the API key for use in OpenCode.

4. Run the `/connect` command and search for **Ollama Cloud**.



```
/connect
```

5. Enter your Ollama Cloud API key.



```
┌ API key

│

│

└ enter
```

6. **Important**: Before using cloud models in OpenCode, you must pull the model information locally:



```
ollama pull gpt-oss:20b-cloud
```

7. Run the `/models` command to select your Ollama Cloud model.



```
/models
```


* * *

### [OpenAI](https://opencode.ai/docs/providers/\#openai)

1. Head over to the [OpenAI Platform console](https://platform.openai.com/api-keys), click **Create new secret key**, and copy the key.

2. Run the `/connect` command and search for OpenAI.



```
/connect
```

3. Enter the API key for the provider.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select the one you want.



```
/models
```


* * *

### [OpenCode Zen](https://opencode.ai/docs/providers/\#opencode-zen-1)

OpenCode Zen is a list of tested and verified models provided by the OpenCode team. [Learn more](https://opencode.ai/docs/zen).

1. Sign in to **[OpenCode Zen](https://opencode.ai/auth)** and click **Create API Key**.

2. Run the `/connect` command and search for **OpenCode Zen**.



```
/connect
```

3. Enter your OpenCode API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Qwen 3 Coder 480B_.



```
/models
```


* * *

### [OpenRouter](https://opencode.ai/docs/providers/\#openrouter)

1. Head over to the [OpenRouter dashboard](https://openrouter.ai/settings/keys), click **Create API Key**, and copy the key.

2. Run the `/connect` command and search for OpenRouter.



```
/connect
```

3. Enter the API key for the provider.



```
┌ API key

│

│

└ enter
```

4. Many OpenRouter models are preloaded by default, run the `/models` command to select the one you want.



```
/models
```









You can also add additional models through your opencode config.



```
{

     "$schema": "https://opencode.ai/config.json",

     "provider": {

       "openrouter": {

         "models": {

           "somecoolnewmodel": {}

         }

       }

     }

}
```

5. You can also customize them through your opencode config. Here’s an example of specifying a provider



```
{

     "$schema": "https://opencode.ai/config.json",

     "provider": {

       "openrouter": {

         "models": {

           "moonshotai/kimi-k2": {

             "options": {

               "provider": {

                 "order": ["baseten"],

                 "allow_fallbacks": false

               }

             }

           }

         }

       }

     }

}
```


* * *

### [SAP AI Core](https://opencode.ai/docs/providers/\#sap-ai-core)

SAP AI Core provides access to 40+ models from OpenAI, Anthropic, Google, Amazon, Meta, Mistral, and AI21 through a unified platform.

1. Go to your [SAP BTP Cockpit](https://account.hana.ondemand.com/), navigate to your SAP AI Core service instance, and create a service key.

2. Run the `/connect` command and search for **SAP AI Core**.



```
/connect
```

3. Enter your service key JSON.



```
┌ Service key

│

│

└ enter
```









Or set the `AICORE_SERVICE_KEY` environment variable:



```
AICORE_SERVICE_KEY='{"clientid":"...","clientsecret":"...","url":"...","serviceurls":{"AI_API_URL":"..."}}' opencode
```









Or add it to your bash profile:



```
export AICORE_SERVICE_KEY='{"clientid":"...","clientsecret":"...","url":"...","serviceurls":{"AI_API_URL":"..."}}'
```

4. Optionally set deployment ID and resource group:



```
AICORE_DEPLOYMENT_ID=your-deployment-id AICORE_RESOURCE_GROUP=your-resource-group opencode
```

5. Run the `/models` command to select from 40+ available models.



```
/models
```


* * *

### [OVHcloud AI Endpoints](https://opencode.ai/docs/providers/\#ovhcloud-ai-endpoints)

1. Head over to the [OVHcloud panel](https://ovh.com/manager). Navigate to the `Public Cloud` section, `AI & Machine Learning` \> `AI Endpoints` and in `API Keys` tab, click **Create a new API key**.

2. Run the `/connect` command and search for **OVHcloud AI Endpoints**.



```
/connect
```

3. Enter your OVHcloud AI Endpoints API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _gpt-oss-120b_.



```
/models
```


* * *

### [Together AI](https://opencode.ai/docs/providers/\#together-ai)

1. Head over to the [Together AI console](https://api.together.ai/), create an account, and click **Add Key**.

2. Run the `/connect` command and search for **Together AI**.



```
/connect
```

3. Enter your Together AI API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Kimi K2 Instruct_.



```
/models
```


* * *

### [Venice AI](https://opencode.ai/docs/providers/\#venice-ai)

1. Head over to the [Venice AI console](https://venice.ai/), create an account, and generate an API key.

2. Run the `/connect` command and search for **Venice AI**.



```
/connect
```

3. Enter your Venice AI API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Llama 3.3 70B_.



```
/models
```


* * *

### [Vercel AI Gateway](https://opencode.ai/docs/providers/\#vercel-ai-gateway)

Vercel AI Gateway lets you access models from OpenAI, Anthropic, Google, xAI, and more through a unified endpoint. Models are offered at list price with no markup.

1. Head over to the [Vercel dashboard](https://vercel.com/), navigate to the **AI Gateway** tab, and click **API keys** to create a new API key.

2. Run the `/connect` command and search for **Vercel AI Gateway**.



```
/connect
```

3. Enter your Vercel AI Gateway API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model.



```
/models
```


You can also customize models through your opencode config. Here’s an example of specifying provider routing order.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "vercel": {

      "models": {

        "anthropic/claude-sonnet-4": {

          "options": {

            "order": ["anthropic", "vertex"]

          }

        }

      }

    }

  }

}
```

Some useful routing options:

| Option | Description |
| --- | --- |
| `order` | Provider sequence to try |
| `only` | Restrict to specific providers |
| `zeroDataRetention` | Only use providers with zero data retention policies |

* * *

### [xAI](https://opencode.ai/docs/providers/\#xai)

1. Head over to the [xAI console](https://console.x.ai/), create an account, and generate an API key.

2. Run the `/connect` command and search for **xAI**.



```
/connect
```

3. Enter your xAI API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _Grok Beta_.



```
/models
```


* * *

### [Z.AI](https://opencode.ai/docs/providers/\#zai)

1. Head over to the [Z.AI API console](https://z.ai/manage-apikey/apikey-list), create an account, and click **Create a new API key**.

2. Run the `/connect` command and search for **Z.AI**.



```
/connect
```









If you are subscribed to the **GLM Coding Plan**, select **Z.AI Coding Plan**.

3. Enter your Z.AI API key.



```
┌ API key

│

│

└ enter
```

4. Run the `/models` command to select a model like _GLM-4.7_.



```
/models
```


* * *

### [ZenMux](https://opencode.ai/docs/providers/\#zenmux)

1. Head over to the [ZenMux dashboard](https://zenmux.ai/settings/keys), click **Create API Key**, and copy the key.

2. Run the `/connect` command and search for ZenMux.



```
/connect
```

3. Enter the API key for the provider.



```
┌ API key

│

│

└ enter
```

4. Many ZenMux models are preloaded by default, run the `/models` command to select the one you want.



```
/models
```









You can also add additional models through your opencode config.



```
{

     "$schema": "https://opencode.ai/config.json",

     "provider": {

       "zenmux": {

         "models": {

           "somecoolnewmodel": {}

         }

       }

     }

}
```


* * *

## [Custom provider](https://opencode.ai/docs/providers/\#custom-provider)

To add any **OpenAI-compatible** provider that’s not listed in the `/connect` command:

1. Run the `/connect` command and scroll down to **Other**.



```
$ /connect




┌  Add credential

│

◆  Select provider

│  ...

│  ● Other

└
```

2. Enter a unique ID for the provider.



```
$ /connect




┌  Add credential

│

◇  Enter provider id

│  myprovider

└
```

3. Enter your API key for the provider.



```
$ /connect




┌  Add credential

│

▲  This only stores a credential for myprovider - you will need to configure it in opencode.json, check the docs for examples.

│

◇  Enter your API key

│  sk-...

└
```

4. Create or update your `opencode.json` file in your project directory:



```
{

     "$schema": "https://opencode.ai/config.json",

     "provider": {

       "myprovider": {

         "npm": "@ai-sdk/openai-compatible",

         "name": "My AI ProviderDisplay Name",

         "options": {

           "baseURL": "https://api.myprovider.com/v1"

         },

         "models": {

           "my-model-name": {

             "name": "My Model Display Name"

           }

         }

       }

     }

}
```









Here are the configuration options:


   - **npm**: AI SDK package to use, `@ai-sdk/openai-compatible` for OpenAI-compatible providers
   - **name**: Display name in UI.
   - **models**: Available models.
   - **options.baseURL**: API endpoint URL.
   - **options.apiKey**: Optionally set the API key, if not using auth.
   - **options.headers**: Optionally set custom headers.

More on the advanced options in the example below.

5. Run the `/models` command and your custom provider and models will appear in the selection list.


* * *

##### [Example](https://opencode.ai/docs/providers/\#example)

Here’s an example setting the `apiKey`, `headers`, and model `limit` options.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "myprovider": {

      "npm": "@ai-sdk/openai-compatible",

      "name": "My AI ProviderDisplay Name",

      "options": {

        "baseURL": "https://api.myprovider.com/v1",

        "apiKey": "{env:ANTHROPIC_API_KEY}",

        "headers": {

          "Authorization": "Bearer custom-token"

        }

      },

      "models": {

        "my-model-name": {

          "name": "My Model Display Name",

          "limit": {

            "context": 200000,

            "output": 65536

          }

        }

      }

    }

  }

}
```

Configuration details:

- **apiKey**: Set using `env` variable syntax, [learn more](https://opencode.ai/docs/config#env-vars).
- **headers**: Custom headers sent with each request.
- **limit.context**: Maximum input tokens the model accepts.
- **limit.output**: Maximum tokens the model can generate.

The `limit` fields allow OpenCode to understand how much context you have left. Standard providers pull these from models.dev automatically.

* * *

## [Troubleshooting](https://opencode.ai/docs/providers/\#troubleshooting)

If you are having trouble with configuring a provider, check the following:

1. **Check the auth setup**: Run `opencode auth list` to see if the credentials
for the provider are added to your config.

This doesn’t apply to providers like Amazon Bedrock, that rely on environment variables for their auth.

2. For custom providers, check the opencode config and:
   - Make sure the provider ID used in the `/connect` command matches the ID in your opencode config.
   - The right npm package is used for the provider. For example, use `@ai-sdk/cerebras` for Cerebras. And for all other OpenAI-compatible providers, use `@ai-sdk/openai-compatible`.
   - Check correct API endpoint is used in the `options.baseURL` field.
