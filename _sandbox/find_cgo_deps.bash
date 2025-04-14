for mod in $(go list -m all); do
    echo "Checking $mod..."
    #go get $mod
    if find "$(go env GOPATH)/pkg/mod/$mod"* -name "*.c" 2>/dev/null | grep -q .; then
        echo "$mod uses CGO"
    fi
done
