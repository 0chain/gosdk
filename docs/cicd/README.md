# Guide to CI/CD using github actions for GOSDK

## Table of Contents
[1. Manual Trigger](#manual-trigger)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.1 For 0proxy from gosdk repo.](#for-0proxy-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.2 For 0box from gosdk repo.](#for-0box-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.3 For 0dns from gosdk repo.](#for-0dns-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.4 For 0block from gosdk repo.](#for-0block-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.5 For 0search from gosdk repo.](#for-0search-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.6 For blobber from gosdk repo.](#for-blobber-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.7 For zboxcli from gosdk repo.](#for-zboxcli-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.8 For zboxmobile from gosdk repo.](#for-zboxmobile-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.9 For zwalletcli from gosdk repo.](#for-zwalletcli-from-gosdk-repo)<br />
[2. Auto Trigger](#auto-trigger)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.1 For production gosdk release from gosdk repo.](#for-production-gosdk-release-from-gosdk-repo)<br />
&nbsp;&nbsp;&nbsp;&nbsp; [1.2 For staging gosdk release gosdk repo.](#for-staging-gosdk-release-from-gosdk-repo)<br />

----
## Manual Trigger

### For 0proxy from gosdk repo
![0proxy](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-0proxy-build.png "UML diagram for 0proxy")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_0PROXY. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/0proxy repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input no/yes for latest tag(Recommended/Default to be "no")<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0proxy](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-0proxy.png "WorkFlow diagram for 0proxy")

----
### For 0box from gosdk repo
![0box](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-0box-build.png "UML diagram for 0box")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_0BOX. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/0box repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input no/yes for latest tag(Recommended/Default to be "no")<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0box](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-0box.png "WorkFlow diagram for 0box")

----
### For 0dns from gosdk repo
![0dns](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-0dns-build.png "UML diagram for 0dns")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_0DNS. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/0dns repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input no/yes for latest tag(Recommended/Default to be "no")<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0dns](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-0dns.png "WorkFlow diagram for 0dns")

----
### For 0block from gosdk repo
![0block](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-0block-build.png "UML diagram for 0block")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_0BLOCK. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/0block repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input no/yes for latest tag(Recommended/Default to be "no")<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0block](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-0block.png "WorkFlow diagram for 0block")

----
### For 0search from gosdk repo
![0search](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-0search-build.png "UML diagram for 0search")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_0SEARCH. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/0search repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input no/yes for latest tag(Recommended/Default to be "no")<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0search](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-0search.png "WorkFlow diagram for 0search")

----
### For blobber from gosdk repo
![blobber](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-blobber-build.png "UML diagram for blobber")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_BLOBBER. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/blobber repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input no/yes for latest tag(Recommended/Default to be "no")<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![blobber](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-blobber.png "WorkFlow diagram for blobber")

----

### For zboxcli from gosdk repo
![zboxcli](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-blobber-build.png "UML diagram for zboxcli")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_ZBOXCLI. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/zboxcli repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input yes For Building With Tag v0.0.0 from master & v1.1.1 from staging<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the latest_tag(Input Tag Example - v1.2.2)<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![zboxcli](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-blobber.png "WorkFlow diagram for zboxcli")

----
### For zboxmobile from gosdk repo
![zboxmobile](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-blobber-build.png "UML diagram for zboxmobile")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_ZBOXMOBILE. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/zboxmobile repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input yes For Building With Tag v0.0.0 from master & v1.1.1 from staging<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the latest_tag(Input Tag Example - v1.2.2)<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![zboxmobile](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-blobber.png "WorkFlow diagram for zboxmobile")

----
### For zwalletcli from gosdk repo
![zwalletcli](https://github.com/0chain/gosdk/blob/master/docs/cicd/trigg-blobber-build.png "UML diagram for zwalletcli")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. TRIGGER_ZWALLETCLI. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/zwalletcli repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input yes For Building With Tag v0.0.0 from master & v1.1.1 from staging<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the latest_tag(Input Tag Example - v1.2.2)<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![zwalletcli](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-blobber.png "WorkFlow diagram for zwalletcli")

----
## Auto Trigger

### For production gosdk release from gosdk repo
![0proxy](https://github.com/0chain/gosdk/blob/master/docs/cicd/build-prod-auto.png "UML diagram for Production")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. GOSDK_RELEASE_PROD. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "master").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/gosdk repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input 'yes' for creating release with tag v0.0.0 & with current release version. If input is 'yes' & branch is 'master', it will also trigger other repositories build with tag 'latest'. If input is 'no', it will only create the release of gosdk with current release version.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0proxy](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-prod.png "WorkFlow diagram for Production")

### For staging gosdk release from gosdk repo
![0proxy](https://github.com/0chain/gosdk/blob/master/docs/cicd/build-stage-auto.png "UML diagram for Staging")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Steps are as follows:-<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 1. Go to the gosdk repository.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2. Click on the Actions to choose the workflow to run.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 3. Choose/Click the workflow i.e. GOSDK_RELEASE_STAGE. Click on the Run workflow.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Select the branch from where to trigger the build(Recommended/Default to be "staging").<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input the branch of 0chain/gosdk repository for creating build.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; NOTE: Input 'yes' for creating release with tag v1.1.1 & with current release version. If input is 'yes' & branch is 'staging', it will also trigger other repositories build with tag 'staging-commit'. If input is 'no', it will only create the release of gosdk with current release version.<br />
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 4. Finally click on the Run Workflow.
![0proxy](https://github.com/0chain/gosdk/blob/master/docs/cicd/workflow-stage.png "WorkFlow diagram for Staging")
