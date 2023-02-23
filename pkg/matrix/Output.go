package matrix

// // actual
// // 0:[{GOARCH {string linux/amd64 }} {version {string go1.17 }}]
// // 1:[{GOARCH {string linux/ppc64le }} {version {string go1.17 }}]
// // 2:[{GOARCH {string linux/s390x }} {version {string go1.17 }} {flags {string -cover -v }}]
// // 3:[{GOARCH {string linux/amd64 }} {version {string go1.18.1 }}]
// // 4:[{GOARCH {string linux/ppc64le }} {version {string go1.18.1 }}]
// // 5:[{GOARCH {string linux/s390x }} {version {string go1.18.1 }}]]

// // wanted
// // 0:[{GOARCH {string linux/amd64 }} {version`` {string go1.17 }}]
// // 1:[{GOARCH {string linux/amd64 }} {version {string go1.18.1 }}]
// // 2:[{GOARCH {string linux/ppc64le }} {version {string go1.17 }} {flags { -cover -v }}]
// // 3:[{GOARCH {string linux/ppc64le }} {version {string go1.18.1 }}]
// // 4:[{GOARCH {string linux/s390x }} {version {string go1.17 }}]
// // 5:[{GOARCH {string linux/s390x }} {version {string go1.18.1 }}]]

// INITIAL COMBINATIONS
// ID 0
// params [{GOARCH {string linux/amd64 [] map[]}} {version {string go1.17 [] map[]}}]
// ID 1
// params [{GOARCH {string linux/ppc64le [] map[]}} {version {string go1.17 [] map[]}}]
// ID 2
// params [{GOARCH {string linux/s390x [] map[]}} {version {string go1.17 [] map[]}}]
// ID 3
// params [{GOARCH {string linux/amd64 [] map[]}} {version {string go1.18.1 [] map[]}}]
// ID 4
// params [{GOARCH {string linux/ppc64le [] map[]}} {version {string go1.18.1 [] map[]}}]
// ID 5
// params [{GOARCH {string linux/s390x [] map[]}} {version {string go1.18.1 [] map[]}}]
