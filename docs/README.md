Getting Started with Slate
------------------------------

### Prerequisites

You're going to need:

 - **Linux or OS X** — Windows may work, but is unsupported.
 - **Ruby, version 2.3.1 or newer**
 - **Bundler** — If Ruby is already installed, but the `bundle` command doesn't work, just run `gem install bundler` in a terminal.

### Developing Docs

1. Run `make dev_docs` to start the slate dev server at http://localhost:4567.
2. Edit/add markdown files in `docs/source/includes`.
3. If you create a new markdown file in `docs/source/includes/`, add it to the includes section of `docs/source/index.html.md`.

Learn more about [editing Slate markdown](https://github.com/lord/slate/wiki/Markdown-Syntax).

### Publishing Docs

1. Run `make publish_docs` to automatically build and push your changes to the gh-pages branch.