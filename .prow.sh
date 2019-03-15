#! /bin/bash -e

# cluster-driver-registrar is not part of any of the current hostpath
# driver deployments, therefore we disable installation and testing of
# those. Instead, the code below runs a custom E2E test suite.
CSI_PROW_DEPLOYMENT=none

. release-tools/prow.sh

# main handles non-E2E testing and cluster installation for us.
if ! main; then
    ret=1
else
    ret=0
fi

if [ "$KUBECONFIG" ]; then
    # We have a cluster. Run our own E2E testing.
    collect_cluster_info
    install_ginkgo
    args=
    if ${CSI_PROW_BUILD_JOB}; then
        # Image was side-loaded into the cluster.
        args=-cluster-driver-registrar-image=csi-cluster-driver-registrar:csiprow
    fi
    run ginkgo -v "./test/e2e" -- -repo-root="$(pwd)" -report-dir "${ARTIFACTS}" $args
fi

exit $ret
