package main

// Run exposes the unexported entry point to the black-box test package so the
// CLI can be driven without building the binary.
var Run = run
