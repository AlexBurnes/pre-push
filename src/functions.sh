# bash script with common definitions and functions
# usage: source "$(dirname "$(readlink -f "$0")")"/functions.sh

set -o errexit
set -o nounset

PWD=$(pwd)
trap cleanup_ SIGINT SIGTERM EXIT
cleanup_() {
    rc=$?
    trap - SIGINT SIGTERM EXIT
    set +e
    [[ "$(type -t cleanup)" == "function" ]] && cleanup
    cd "${PWD}"
    exit $rc
}

if [[ -t 2 ]] && [[ -z "${NO_COLOR-}" ]] && [[ "${TERM-}" != "dumb" ]]; then
    clre='\e[0m' black='\e[30m' red='\e[31m' green='\e[32m' yellow='\e[33m' 
    blue='\e[34m' magenta='\e[35m' cyan='\e[36m' gray='\e[37m' white='\e[38m' 
    bold='\e[1m' blink='\e[5m]'
else
    clre='' black='' red='' green='' yellow='' 
    blue='' magenta='' cyan='' gray='' white='' 
    bold='' blink=''
fi

SCRIPT=$(readlink -f "$0")
SCRIPT_NAME=$(basename "${SCRIPT}")
SCRIPT_PATH=$(dirname "${SCRIPT}")
PROJECT_DIR=$(dirname "${SCRIPT_PATH}")
[[ -f "${PROJECT_DIR}"/.project ]] && source "${PROJECT_DIR}"/.project
