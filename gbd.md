# appc Binary Discovery (abd)

## Overview 

abd, appc Binary Discovery, defines a general framework for converting a human-readable string to a downloadable artefact URI.
It supports a federated namespace, but with configurable overrides and multiple discovery mechanisms.
It is transport-agnostic and provides a simple and extensible interface.

## Motivation

Create a clean layer for naming and transport that can be used by different container specifications and runtimes. We take inspiration from existing systems such as rpm, apt, appc, Docker, and others. The spec strives to use known best practices and existing protocols when possible.

## abd resolution process

The ABD resolution process consists of two steps:

1) <identifier> + <label selectors> --> <metadata-fetch-strategy> --> <ABD metadata blob>
2) <ABD metadata blob> --> <client filtering w/label selectors> --> <suitable artefact mirrors>

In the `abd` examples below, step 1) is implemented by `abd discover` and steps 1+2) are implemented by `abd mirrors`.

## abd tool

We illustrate ABD with a simple command-line tool.

`abd discover` takes `<identifier>` and set of `<label selectors>` (arbitrary key-value pairs), applies an appropriate metadata-fetch-strategy (chosen based on configuration described later in this document), and yields a blob of JSON in the ABD Metadata Format.

```
abd discover com.coreos.etcd,content-type=aci,os=linux,arch=amd64
```

`abd mirrors` will retrieve the metadata, perform label filtering, and extract a suitable list of URIs (mirrors) from which the corresponding artifact can be retrieved.

```
	abd mirrors com.coreos.etcd,content-type=aci,os=linux,arch=amd64
```

`abd fetch` will actually retrieve the artiefact from one of the mirrors:

```
abd fetch com.coreos.etcd,content-type=aci,os=linux,arch=amd64
```

Application examples (these could either fork out to `abd` or implement the specified semantics internally):

```
    rkt fetch com.coreos.etcd 	# implies content-type=aci,os=linux,arch=amd64
    docker fetch com.coreos.etcd 	# implies content-type=docker,os=linux,arch=amd64
```

## ABD Metadata Format 

Applying a metadata-fetch-strategy to an `<identifier>` + `<label selectors>` returns a blob of JSON in ABD Metadata Format: 
- json blob of (name, labels, mirrors)
- labels are arbitrary key-value pairs (same as label selectors)
- mirrors is a list of URIs which MUST correspond to the same artefact

** OPEN QUESTION: perhaps this is just TUF, and we embed all of the abd stuff inside the TUF metadata? ** 
** OPEN QUESTION: do we want to define an interface for fetching URIs in the mirrors? This could be used by `abd fetch`. For example, we would just pass `http(s)://` to `wget`, or `hdfs://` to `hadoop fs get`, and so on **

Example ABD Metadata Format blob:
```
"metadata": {"io.abd.metadata":
[
 {
  "name": "com.coreos.etcd",
  "labels": {
   "version": "1.0.0",
   "arch": "amd64",
   "os": "linux",
   "content-type": "application-binary/aci"
  },
  "mirrors": [
    "https://github.com.../etcd-linux-amd64-1.0.0.aci",
  ]
 }
],
[
 {
  "name": "com.coreos.etcd",
  "labels": {
   "version": "1.0.0",
   "arch": "amd64",
   "os": "linux",
   "content-type": "docker",
  },
  "mirrors": [
    "docker://quay.io/coreos/etcd:1.0.0",
    "docker://gcr.io/coreos/etcd:1.0.0"
  ]
 }
]}
```

## ABD Client Configuration (metadata-fetch-strategy configuration)
The configuration defines how the abd tool chooses which metadata-fetch-strategy to use, given a certain identifier+labels.

The configuration format is loosely inspired by Debian apt repositories' [sources.list configuration][apt-sources-list]. 
- Lexically-ordered configuration files, each containing a single strategy configuration
- Given an identifier, all strategies with prefixes that match that identifier are tried in order
- Try both metadata retrieval AND artefact retrieval for each strategy before moving on to the next
- Each configuration has two fields prescribed by abd: `prefix` and `strategy`
- All other fields defined in the configuration are passed unaltered to the strategy

Below are some example configurations for various use cases.

#### Scenario: default configuration: use https+dns for all

** OPEN QUESTION: this is what we want the "default" abd behaviour to be; but perhaps rather than having it implicit, we could keep it explicit, i.e. set the expectation that it is actually shipped as a default configuration file with abd**

```
$ ls /usr/lib/abd/sources.list.d
zz-default.conf
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
 "prefix": "*",
 "strategy": "io.abd.https-dns"
}
```

#### Scenario: local (cache) + http+dns fallback

```
$ ls /usr/lib/abd/sources.list.d/
10-local.conf 
zz-default.conf
$ cat /usr/lib/abd/sources.list.d/10-local.conf
{
    "prefix": "*",
    "strategy": "io.abd.local",
    "storage-path": "/var/abd/"
}
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
    "prefix": "*",
    "strategy": "io.abd.https-dns"
}
```

#### Scenario: block all fetching

```
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
    "prefix": "*",
    "strategy": "io.abd.noop"
}
```

#### Scenario: use nfs for everything under com.coreos*, http-dns for everything else

```
$ ls /usr/lib/abd/sources.list.d/
10-local-nfs-coreos.conf 
zz-default.conf
$ cat /usr/lib/abd/sources.list.d/10-local-nfs-coreos.conf
 {
    "prefix": "com.coreos*",
    "strategy": "io.abd.nfs",
    "fetch-uri": "nfs:///share/aci-cache"
}
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
    "prefix": "*",
    "strategy": "io.abd.https-dns"
}
```

With this configuration,


abd fetch com.coreos.etcd

results in looking for metadata at 

nfs:///share/aci-cache/abd-index/com.coreos.etcd

but 
abd fetch com.oracle

results in looking for metadata at 

https://oracle.com/.well-known/abd-index/com.oracle

## ABD metadata fetch strategies 

ABD defines a set of well-known metadata fetch strategies and an interface for implementing additional ones.
(The well-known metadata fetch strategies conform to the same interface, and could simply be implemented as "plugins that ship by default").

### ABD metadata fetch strategy interface

ABD defines a simple fork-exec plugin interface for metadata-fetch-strategy. The binary receives the referring strategy configuration JSON on stdin, and the abd identifier + labels as arguments, and outputs a JSON blob in the ABD Metadata Format to stdout:

```
   cat 10-local-nfs-coreos.conf | <abd-plugin> ${identifier} ${labels}...
```

### Well-known strategies (i.e., would ship with `abd`)

#### `io.abd.https-dns`

  This strategy looks for metadata at [well-known URIs][well-known-rfc] over HTTPS, walking up the DNS tree (similar to [appc meta discovery][appc-meta-discovery]).
 It passes labels as query parameters (server can optionally use these query parameters to provide server-side filtering as an optimisation).

 For example, given the identifier `com.coreos.etcd`, first attempt:

    wget "https://etcd.coreos.com/.well-known/abd-index/com.coreos.etcd?label1=foo&label2=bar"

 Assuming this would fail because of no DNS entry, next attempt:

    wget "https://coreos.com/.well-known/abd-index/com.coreos.etcd?label1=foo&label2=bar"

 this succeeds and returns the appropriate metadata blob.

#### `io.abd.local`

   Parameters: `storage-path`

  This is a very simple strategy intended to facilitate someone just having a directory of artefacts on the local system.
  It uses a built-in template to (internally) generate a metadata blob - the template is substituted and returned to abd:

"io.abd.metadata":
[
 {
  "name": "{abd_identifier}",
  "mirrors": [
    "{storage-path}/{abd-identifier}",
  ]
 }
]

#### `io.abd.noop`

  This strategy always fails; it is intended to be used to block retrievals for a certain prefix. It returns an empty metadata blob (or, equivalently, one with an empty list of mirrors).

#### `io.abd.nfs`

  Parameters: `fetch-uri`

  This strategy attempts to retrieve metadata blob from `nfs:///${fetch-uri}/abd-index/${abd_identifier}`


[well-known-rfc]: https://tools.ietf.org/html/rfc5785
[apt-sources-list]: http://manpages.debian.org/cgi-bin/man.cgi?sektion=5&query=sources.list&apropos=0&manpath=sid&locale=en
[appc-meta-discovery]: https://github.com/appc/spec/blob/master/spec/discovery.md#meta-discovery


## Addendum: ABD and appc

If ABD is implemented (as an independent specification), the authors of appc propose that it replace the existing discovery section of appc. 

There are a number of outstanding issues/questions around appc discovery that this should resolve:
- Most notably, Discovery should be reframed as a selectable set of discovery strategies
- use TXT dns records for discovery (this could be implemented as a plugin)
- where do arch/os/version labels come from (these would no longer be special labels, and simple discovery would be scrapped)
- use well-known URIs (this is suggested default in new proposal)
- adopting TUF (part of this proposal)
- supporting S3-backed repositories (should be easy to implement with new framework)

