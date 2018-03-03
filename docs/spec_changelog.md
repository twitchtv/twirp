---
id: "spec_changelog"
title: "Twirp Wire Protocol Changelog"
sidebar_label: "Changelog"
---

This document lists changes to the Twirp specification.

## Changed in v6

### URL scheme

In [v5](./PROTOCOL.md), URLs followed this format:

```bnf
**URL ::= Base-URL "/twirp/" [ Package "." ] Service "/" Method**
```

Version 6 changes this format to remove the mandatory `"/twirp"` prefix:

```bnf
**URL ::= Base-URL "/" [ Package "." ] Service "/" Method**
```

Also, `Base-URL` can now contain a path component - in other words, it's legal
to set any prefix you like.

If you loved the old `/twirp` prefix, you can still use it by using a base URL
that ends with `/twirp`. You're no longer forced into use it, however.

The "/twirp/" prefix is no longer required for three reasons:

 - Trademark concerns: some very large organizations don't want to
   take any legal risks and are concerned that "twirp" could become
   trademarked.
 - Feels like advertising: To some users, putting "twirp" in all your
   routes feels like it's just supposed to pump Twirp's brand, and
   provides no value back to users.
 - Homophonous with "twerp": In some Very Serious settings (like
   government websites), it's not okay that "twirp" sounds like
   "twerp", which means something like "insignificant pest."
