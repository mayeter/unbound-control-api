#!/bin/bash

## This is an installation script for Unbound Control API.
## It always install the latest release of the project.

## Here are the defaults
Binary_File_Path="/usr/bin/unbound-control-api"
Config_Dir="/etc/unbound-control-api/"
Service_User="unbound-control-api"
SystemD_Path=$(pkg-config systemd --variable=systemdsystemunitdir)

## New release url should be given as parameter
NewReleaseURL=$1

###################################################################################
############################### Definition of Functions ###########################
###################################################################################
function checkCurrentDir() {
    Required_Files=["api.conf","unbound-control-api.service","unbound-control-api"]
    for file in $Required_Files
    do
      if [[ ! -f ./$file ]]
      then
        echo "Some of required files are missing. Please look it up from README.md"
        exit 10
      fi
    done
}

function unzipNewRelease() {
    curl https://$NewReleaseURL --output ./latest.zip
    unzip ./latest.zip
    cd latest
}

function createConfigDir() {
    if [[ ! -f $Binary_File_Path ]]
    then
      mkdir $Config_Dir
    fi
    cp release-package/api.conf $Config_Dir/api.conf
    chown -R $Service_User:$Service_User $Config_Dir
    chmod -R 644 $Config_Dir
}

function createSystemdUnitFile() {
  if [[ ! -f $SystemD_Path/unbound-control-api.service ]]
  then
    cp release-package/unbound-control-api.service $SystemD_Path/unbound-control-api.service
    systemctl daemon-reload
    systemctl enable unbound-control-api
  fi
}

function copyBinaryFile() {
    cp release-package/unbound-control-api $Binary_File_Path
}

function killOldProcess () {
  processID=$(lsof /usr/sbin/unbound | tail -n 1 | awk '{print $2}')
  if [[ -z $processID ]]
  then
    systemctl stop unbound-control-api
    #kill -9 "$processID"
  fi
}

function runNewProcess() {
  systemctl start unbound-control-api
  lsof $Binary_File_Path
}



###################################################################################
############################### Execution of Functions ############################
###################################################################################


if [[ -z $NewReleaseURL ]]
then
  echo "A url is not provided!"
  echo "Installing from current directory"
else
  echo "Downloading newer version from the URL specified!"
  unzipNewRelease
fi

## These two will be executed in any case
checkCurrentDir
copyBinaryFile


## Lets learn about current state
## Is it installed before?
if [[ ! -f $Binary_File_Path ]]
## It is not installed before
then
  createConfigDir
  createSystemdUnitFile
  runNewProcess


## It's installed previously
else
  ## Is it working right now?
  if lsof $Binary_File_Path
  ## Yes, it is active
  then
    killOldProcess
    runNewProcess

  ## No, it's not active
  else
    runNewProcess
  fi

fi
