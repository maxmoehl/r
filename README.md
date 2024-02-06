# r - Redirect

A simple service to build shortcuts. Recommended to combine with tools like
[Keyword Search](https://apps.apple.com/de/app/keyword-search/id1558453954) or
[Kagi Custom Bangs](https://help.kagi.com/kagi/features/bangs.html#custom-bangs).

## Installation

```sh
go install moehl.dev/r@latest
```

## Configuration

_Note: Although this JSON5, the application only understands plain JSON._

A request is a space separated list of strings supplied as the `to` query parameter. For example
`gh r` or `gh cf r`.

```json5
{
  // The configuration is a tree-like structure, all nodes have the same structure.

  // URL is returned as a redirect if the request matches exactly this node:
  "url": "https://example.com",

  // Template is a catch-all if no child matches, a single '%s' can be put in to url-encode any
  // excess arguments (using url.PathEscape and fmt.Sprintf):
  "template": "https://duckduckgo.com/?q=%s",

  // Children are matched against the next argument:
  "children": {
    // Shortcut to reach GitHub
    "gh" {
      // Base URL:
      "url": "https://github.com",
      // Default is to search
      "template": "https://github.com/search?q=%s",
      // And some custom shortcuts
      "children": {
        "r": {
          "url": "https://github.com/maxmoehl/r"
        },
        "cf": {
          "url": "https://github.com/cloudfoundry",
          "children": {
            "r": {
              "url": "https://github.com/cloudfoundry/gorouter"
            }
          }
        }
      }
    }
  }
}
```

Depending on your use-case there can be a lot of repetition. I'm planning to add a feature which
makes building the URL additive, so the `gh` node could contain only the base path and the `cf`
node would only contain `/cloudfoundry`.

For most cases the templating should be strong enough too avoid repititon.
