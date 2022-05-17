#!/bin/bash

set -e

# interpretation is only here for the time being until I can properly set up integration tests for it
VR_PACKAGES=`go list ./... | grep -v interpretation | grep -v recog | grep -v response | grep -v view`

REPO_ROOT="$GOPATH/src/github.com/gavinturner/vinylretails"
TEST_COACH_PATH="$REPO_ROOT/util/test-coach.rb"

print_help() {
    echo "Run all tests:                  ./test.sh"
    echo "Run all tests for package api:  ./test.sh api"
    echo "Run all unit tests:             ./test.sh --unit [<package>]"
    echo "Run all lint tests:             ./test.sh --lint [<package>]"
    echo "Run a full suite of all tests:  ./test.sh --suite [<package>]"
    echo "Run a specific Go test:         ./test.sh --func <package> <test>"
}

run_all() {
	go test -tags="unit_test integration_test" "$@"
}

run_unit_all() {
	go test -short -tags=unit_test "$@"
}

run_integration_all() {
    go test -tags=integration_test "$@"
}


run_go_vet() {
    echo "go vet"
    go vet "$@"
}

run_check_licenses() {
    echo "Checking licenses..."
    ruby "$REPO_ROOT/util/find_licenses.rb" -f
    if [[ $(grep "GNU GPL" licenses.txt) ]]; then
	echo "We're not allowed to have GPL licensed code"
	exit 1
    else
	echo "Licenses good"
    fi
}

run_golint() {
    # list exceptions here. (official golint docs say "to ignore some checks, use tools like grep")
    echo "golint <package(s)>"
    $GOPATH/bin/golint "$@" \
	| grep -v -E "exported (const|function|method|type|var) [^ ]+ should have comment" \
	| grep -v -E "if block ends with a return statement, so drop this else and outdent its block" \
	| grep -v -E "comment on exported (const|function|method|type|var) [^ ]+ should be of the form" \
	| grep -v -E "don't use underscores in Go names" \
	| grep -v -E "(func|struct field) \S*(Api|Ip|Sql|Url|Uuid)\S* should be \S*(API|IP|SQL|URL|UUID)\S*" \
	| grep -v -E "should omit 2nd value from range" \
	| grep -v -E "don't use an underscore in package name" \
	| grep -v -E "\S*Id\S* should be \S*ID\S*" || :
}

header() {
    echo -e "\033[33m$1\033[0m"
}

run_specific_test() {
    PACKAGE=$1
    TEST=$2
    go test -v --tags="unit_test integration_test" github.com/gavinturner/vinylretailers/$PACKAGE -run $TEST
}

run_suite() {
    header "0. Compiling project..."
    echo "go build ./..."
    go build ./...

    header "1. Checking licenses..."
    run_check_licenses

    header "2. Formatting project..."
    echo "go fmt"
    go fmt $VR_PACKAGES

    header "3. Vetting project..."
    run_go_vet $VR_PACKAGES

    header "4. Linting the project..."
    run_golint $VR_PACKAGES

    if [ -x $TEST_COACH_PATH ]; then
	header "5. checking over test files..."
	echo $TEST_COACH_PATH
	$TEST_COACH_PATH
    fi

    header "6. Running unit tests for project..."
    echo "./test.sh --unit"
    ./test.sh --unit

    header "7. Running integration tests on project..."
    echo "./test.sh"
	  ./test.sh --integration

    header "DONE"
}

run_suite_pack() {
    header "1. Compiling project ..."
    echo "make b"
    make b

    header "2. Formatting package $1 ..."
    echo "go fmt github.com/gavinturner/vinylretailers/$1"
    go fmt github.com/gavinturner/vinylretailers/$1

    header "3. Vetting package $1 ..."
    run_go_vet github.com/gavinturner/vinylretailers/$1

    header "4. Linting package $1 ..."
    run_golint github.com/gavinturner/vinylretailers/$1

    if [ -x $TEST_COACH_PATH ]; then
	header "5. checking over test files in package $1 ..."
	echo $TEST_COACH_PATH
	$TEST_COACH_PATH $1
    fi

    header "6. Running unit tests for package $1 ..."
    echo "./test --unit $1"
    ./test.sh --unit $1

    header "7. Running integration tests for package $1 ..."
    echo "./test.sh $1"
    ./test.sh --integration $1

    header "DONE"
}

run_pack() {
    go test -v -tags="unit_test integration_test" "github.com/gavinturner/vinylretailers/$1"
}

run_unit_pack() {
	go test -short -tags=unit_test -v "github.com/gavinturner/vinylretailers/$1"
}

run_integration_pack() {
	go test -tags=integration_test -v "github.com/gavinturner/vinylretailers/$1"
}

case "$1" in
    "")
	echo "Running all tests"
	run_all $VR_PACKAGES
	;;
    "--help")
	print_help
	;;
    "--suite")
	case "$2" in
	    "")
		echo "Running test suite"
		run_suite
		;;
	    *)
		echo "Running test suite for package $2"
		run_suite_pack "$2"
		;;
	esac
	;;
    "--unit")
	case "$2" in
	    "")
		echo "Running all unit tests"
		run_unit_all $VR_PACKAGES
		;;
	    *)
		echo "Running all unit tests for package $2"
		run_unit_pack "$2"
		;;
	esac
	;;
    "--integration")
	case "$2" in
	    "")
		echo "Running all integration tests"
		run_integration_all $VR_PACKAGES
		;;
	    *)
		echo "Running all integrationg tests for package $2"
		run_integration_pack "$2"
		;;
	esac
	;;
    "--vet")
	case "$2" in
	    "")
		echo "Vetting project"
		run_go_vet $VR_PACKAGES
		;;
	    *)
		echo "Vetting package $2"
		run_go_vet "github.com/gavinturner/vinylretailers/$2"
		;;
	esac
	;;
    "--lint")
	case "$2" in
	    "")
		echo "Linting project"
		run_golint $VR_PACKAGES
		;;
	    *)
		echo "Linting package $2"
		run_golint "$2"
		;;
	esac
	;;
    "--func")
	case "$2" in
	    "")
		echo "Usage: ./test.sh --func <package> <test>"
		;;
	    *)
		case "$3" in
		    "")
			echo "Usage: ./test.sh --func <package> <test>"
			;;
		    *)
			run_specific_test $2 $3
			;;
		esac
		;;
	esac
	;;
    *)
	echo "Running all tests for package $1"
	run_pack "$1"
	;;
esac
