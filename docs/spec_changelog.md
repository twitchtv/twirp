---
id: "spec_v5"
title: "Twirp Wire Protocol Changelog"
sidebar_label: "Changelog"
---

This document lists changes to the Twirp specification.

## Changed in v6

### URL scheme

In [v5](./PROTOCOL.md), URLs followed this format:

**URL ::= Base-URL "/twirp/" [ Package "." ] Service "/" Method**

In v6, the "/twirp/" prefix is removed:

**URL ::= Base-URL "/" [ Package "." ] Service "/" Method**

The "/twirp/" prefix is removed for three reasons:

 - Trademark concerns: some very large organizations don't want to
   take any legal risks and are concerned that "twirp" could become
   trademarked.
 - Feels like advertising: To some users, putting "twirp" in all your
   routes feels like it's just supposed to pump Twirp's brand, and
   provides no value back to users.
 - Homophonous with "twerp": In some Very Serious settings (like
   government websites), it's not okay that "twirp" sounds like
   "twerp", which means something like "insignificant pest."
