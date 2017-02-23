#!/bin/env python
"""coverage.py

This script is for checking the code coverage of unit tests in the
oshinko-rest project. It is meant to be invoked from the top level of the
repository.

Example invocation:

    $ tools/coverage.py -h

"""
import argparse
import copy
import re
import subprocess


oshinko_repo = 'github.com/radanalyticsio/oshinko-cli/rest/'
oshinko_test_package = oshinko_repo + 'tests/unit'
coverage_packages = [
    'handlers',
    'helpers/containers',
    'helpers/deploymentconfigs',
    'helpers/errors',
    'helpers/info',
    'helpers/podtemplates',
    'helpers/services',
    'helpers/uuid',
    'version',
]


def main(args):
    def run_and_print(cmd):
        proc = subprocess.Popen(cmd,
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        match = re.search('[0-9]{1,3}\.[0-9]%', proc.stdout.read())
        if match is not None:
            print('   ' + match.group(0))
        else:
            print('   unknown')

    print('starting coverage scan')
    base_cmd = ['go', 'test']
    if args.coverprofile is not None:
        base_cmd = base_cmd + ['-coverprofile', args.coverprofile]
    if args.individual is True:
        for pkg in coverage_packages:
            print(' - scanning ' + pkg)
            cmd = base_cmd + ['-coverpkg', oshinko_repo+pkg,
                              oshinko_test_package]
            run_and_print(cmd)
    else:
        print(' - scanning all packages')
        pkg_list = ','.join([oshinko_repo+p for p in coverage_packages])
        cmd = base_cmd + ['-coverpkg', pkg_list, oshinko_test_package]
        run_and_print(cmd)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Run coverage analysis.')
    parser.add_argument('-i', '--individual', dest='individual',
                        action='store_true',
                        help='Print coverage analysis for each package.')
    parser.add_argument('-c', '--coverprofile', dest='coverprofile',
                        action='store',
                        help='Write coverage profile to this file.')
    args = parser.parse_args()
    main(args)
