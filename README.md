Agnostic Raw Data (ARD) for Go
==============================

This library is [also implemented in Python](https://github.com/tliron/python-ard).

And check out the [ardconv](https://github.com/tliron/ardconv) ARD conversion tool.

What is "agnostic raw data"?

### Agnostic

ARD comprises data types that are "agnostic", meaning that they can be trivially used by
practically any programming language, stored in practically any database, and can also be
transmitted in a wide variety of formats.

The following data types are supported:

* strings (Unicode)
* byte arrays
* signed integers
* unsigned integers
* floats
* booleans
* nulls

As well as two nestable structures:

* lists
* maps (unordered)

Note that map keys *do not have to be strings* and indeed can be arbitrarily complex. Such keys
might be impossible to use in hashtable implementations in some programming languages. In such
cases maps can be stored as lists of key/value tuples.

### Raw

Data validation is out of scope for ARD. There's no schema. The idea is to support *arbitrary*
data of any structure and size. Once the ARD is made available other layers can validate its
structure and otherwise process the values.

This library does support such schema validation via conversion to Go structs using a
[reflector](reflection.go).

### Data

This is about *data* as opposed to the *representation of data*. What's the difference? ARD does
not define *how* the data is stored or transmitted. Thus ARD in itself is not concerned with the
endiannes or precision of integers and floats, and also not concerned with character encodings
(compare the Unicode standard for data vs. the UTF-8 standard for encoding that data).

ARD and Representation Formats
------------------------------

### CBOR and MessagePack

[CBOR](https://cbor.io/) and [MessagePack](https://msgpack.org/) support everything! Though note
that they are not human-readable.

### YAML

YAML supports a rich set of primitive types (when it includes the common
[JSON schema](https://yaml.org/spec/1.2/spec.html#id2803231)), so most ARD will survive a round
trip to YAML.

YAML, however, does not distinguish between signed and unsigned integers.

Byte arrays can also be problematic. Some parsers support the optional
[`!!binary`](https://yaml.org/type/binary.html) type, but others may not. Encoded strings (e.g.
using Base64) can be used instead to ensure portability.

Also note that some YAML 1.1 implementations support ordered maps
([`!!omap`](https://yaml.org/type/omap.html) vs. `!!map`). These will lose their order when
converted to ARD, so it's best to standardize on arbitrary order (`!!map`). YAML 1.2 does not
support `!!omap` by default, so this use case may become less and less common.

### JSON

JSON can be read into ARD. However, because JSON has fewer types and more limitations than YAML
(no signed and unsigned integers, only floats; map keys can only be strings), ARD will lose quite a
bit of type information when translated into JSON.

We overcome this challenge by extending JSON with some conventions for encoding extra types.
See [our conventions here](cjson.go) or
[in the Python ARD library](https://github.com/tliron/python-ard/blob/main/ard/cjson.py).

### XML

XML does not have a type system. Arbitrary XML cannot be parsed into ARD. 

However, we support [certain conventions](xml.go) that enforce such compatibility.

ARD and Programming Languages
-----------------------------

### Go

Unfortunately, the most popular Go YAML parser does not easily support arbitrarily complex keys
(see this [issue](https://github.com/go-yaml/yaml/issues/502)). We provide an independent library,
[yamlkeys](https://github.com/tliron/yamlkeys), to make this easier.

### Python

Likewise, the Python [ruamel.yaml](https://yaml.readthedocs.io) parser does not easily support
arbitrarily complex keys. We solve this by extending ruamel.yaml in our
[Python ARD library](https://github.com/tliron/python-ard).

### JavaScript

See the discussion of JSON, above (JSON stands for "JavaScript Object Notation"). A
straightforward way to work with ARD in JavaScript is via our ARD-compatible JSON conventions.
However, it may also be possible to create a library of classes to support ARD features.
