# LENS-SLACK-BOT
This is a slack bot which can take commands from the slack interface.
It alerts the user on the following:
- New proposals on the chain
- When user didn't vote on the proposal

 
## Install Prerequisites

    Go 13.x+
    Ubuntu 16.04 +

The applications required for this bot are 
* Slack
* SQLite

## Slack Installation

**Steps to install slack:**

To install Slack on Linux, follow these steps:
* Open the terminal on your Linux system
* Enter the following command to install the Slack snap package:

           sudo snap install slack --classic

* Wait for the installation to complete.

* Once the installation is complete, you can launch Slack from the applications menu or by running the slack command in the terminal.

To install Slack on Windows, follow these steps:

* Go to the Slack download page in your web browser.
* Click the Download button to download the Slack installer.
* Once the download is complete, open the Slack installer by double-clicking the downloaded file.
* Follow the prompts in the installer to complete the installation process.
* When the installation is complete, open Slack and sign in with your workspace credentials.

## SQLite Installation

**Steps to install SQLite:**

To install SQLite on Linux, follow these steps:

* Open the terminal on your Linux system.

* Update the package manager by entering the following command and pressing Enter:

        sudo apt update


* Install SQLite by entering the following command and pressing Enter:

        sudo apt install sqlite3

* This will install SQLite on your system.

* Test the SQLite installation by running the sqlite3 command in the terminal. For example, type the following command and press Enter:

        sqlite3 --version

    If the SQLite version is displayed, the installation was successful.

## Building a Slack Bot

**Steps to build a slack bot**

* Create a Slack workspace: If you don't already have a Slack workspace, create one by visiting the Slack website and following the instructions.

* Create a Slack app: Once you have a workspace, go to the Slack API website and click the Create New App button. Give your app a name and select the workspace where you want to use it.

* Add a bot user: In your app's settings, click the Bot Users tab and click the Add a Bot User button. Give your bot a display name and default username, then click Add Bot User.

* Install the app: In your app's settings, click the Install App tab and click the Install App to Workspace button. Follow the prompts to authorize the app and install it in your workspace.

* Obtain the bot token: After the app is installed, you can obtain the bot token from the OAuth & Permissions tab in your app's settings. Copy the bot token to use it in your bot's code.

*  Create a bot script: Script can be created using any required programming language which can build the bots core functions using databases and various slack function

*    Test the bot: Once your bot is up and running, test it in your Slack workspace by sending it a message or invoking a command.

## Configure the following variables in config.toml
 Our config file mainly consists of three main components required to run a slack bot
 
 * Bot Token:
            A token used by a bot to authenticate with the Slack API and perform actions on behalf of the bot user.
 * Socket Token:
     A token used by a Slack app to establish a connection to the Slack API using WebSocket protocol, enabling real-time communication.
 * Channel ID:
     A unique identifier for a Slack channel that can be used in API requests to interact with the channel 
     
 Steps to get these tokens
 
 ### **Bot Token**
  
  Bot token can be acquired from the api.slack website by following these steps
  * Go to the website and find the "Your Apps" option. (Bot Token can only be acquired after building a bot/app)
  * Find the specific bot required and click on it.
  * Under the "Features" Menu you will find "OAuth&Permissions". Under this section You can find the bot token under "Oauth tokens for your tokens"
  * Here's a similar example 
  
          xoxb-5192976590416-5170323504898-Z5RwjjXxgyJbTiuktkYTFj
  ### **Socket Token**
  
  Socket token can be acquired from the api.slack website by following these steps
  * Go to the website and find the "Your Apps" option. (Socket Token can only be acquired after building a bot/app)
  * Find the specific bot required and click on it.
  * Under the "Settings" Menu you will find "Basic Information". Under this section You can find and choose the socket token under "App level tokens"
  * Here's a similar example 
  
          xapp-1-A055BTNEYAD-5162715820838-57b8f785479baed7c1a2188c719e1c13d296effebc
          
### Channel ID

* Open your specific workspace on web.
* Go your channel tab and take a note of your web address which will be similar to below
            
        app.slack.com/client/H055NUQHDB8/C0551NPBGTW/

* In that specific web address You will be able to find an ID similar to below example from above reference

        " C0551NPBGTW "
        
## Here is the list of available alerts and Slack bot commands

* Alert on new proposals
* Alert for every 2 hours if the voting period ends in less than 24 hours 
   
   ### List of avaliable telegram commands

To get response from these commands you can just @<bot-name> <command_name> from your slack workspace/channel. You will be getting response to your slack workspace/channel based on the bot token/channel ID you have configured in config.toml

    register : registers the validator using chain id and validator address
    create-key : Create a new account with key name
    vote : vote on proposals 
         !note!: keys are an attribute which are added and need to be funded in order for the voting to progress. The grantee must give the voter authorisation before the voting can proceed 
    list-keys : List of all the key names
    list-all : List of all registered chain id & validators address
