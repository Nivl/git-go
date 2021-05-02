[Replace All mentions of CVE-ID by the ID of the CVE]
[Replace All mentions of AFFECTED-PACKAGE-IMPORT-PATH by the affected package, like github.com/gogo/protobuf]
[Replace All mentions of AFFECTED-PACKAGE-FIXED-VERSION by the version that fixes the CVE]
[See Example: https://github.com/Nivl/git-go/pull/113]

### Core info

- Link to affected dep: https://AFFECTED-PACKAGE-IMPORT-PATH
- CVSS SCORE:
- [CVE-ID](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-ID)
- [National Vulnerability Database](https://nvd.nist.gov/vuln/detail/CVE-ID)

### Severity & impact on go-git

[Explain what this CVE does, how it can be used, and how it affects the project]

### Explanation

[What are the affected depenencies, which root dependency is affected, what are the possible fixes]

### Research

[Part of dependency graphs that show how the nested dependencies are affected by the graph]

CVE in: `AFFECTED-PACKAGE-IMPORT-PATH` < AFFECTED-PACKAGE-FIXED-VERSION

Affected dependencies:

```
❯ go mod graph | grep " AFFECTED-PACKAGE-IMPORT-PATH"
// paste output
```

[add more graph if needed, until reaching a direct dependency. Only copy the part that shows]

### Nancy report

```
// Copy Nancy reports here
```
