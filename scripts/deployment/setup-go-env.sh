#!/bin/bash
# Go Environment Setup Script for Blytz MVP
# This script sets up the proper Go environment for development

export GOTOOLCHAIN=go1.25.0
export GOPATH="/home/sas/go"

echo "✅ Go environment configured:"
echo "   PATH: $PATH"
echo "   GOROOT: $GOROOT"
echo "   GOPATH: $GOPATH"
echo
echo "Go version: $(go version)"
echo
echo "🚀 Environment ready for microservices development!"