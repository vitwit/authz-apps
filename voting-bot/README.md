# LENS-SLACK-BOT
This is a slack bot which can take commands from the slack interface.
It alerts the user on the following:
- New proposals on the chain
- When user didn't vote on the proposal

 
## Install Prerequisites

    Go 1.9 or higher
    Ubuntu 16.04 +

The applications required for this bot are 
* Slack
* SQLite

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
    * Before installing it into the work space try to add scopes which are located under the OAuth & Permissions.some of the basic scopes which can be added are app_mention:read channel:history,channel:read,chat:write,im:history ,user:read
    * If required events can also be initialised to make the bot more efficient .Some of the basic events are app_mention,message.channel,message.im,im_history_changed.

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
  
          xoxb-XXXXXXXXXXXX
  ### **Socket Token**
  
  Socket token can be acquired from the api.slack website by following these steps
  * Go to the website and find the "Your Apps" option. (Socket Token can only be acquired after building a bot/app)
  * Find the specific bot required and click on it.
  * Under the "Settings" Menu you will find "Basic Information". Under this section You can find and choose the socket token under "App level tokens"
  * Here's a similar example 
  
          xapp-1-XXXXXXXXXXXXXX
          
### Channel ID

* Open your specific workspace on web.
* Go your channel tab and take a note of your web address which will be similar to below
            
        app.slack.com/client/XYZ/CABC/

* In that specific web address You will be able to find an ID similar to below example from above reference

        "CABC"
        
## Here is the list of available alerts and Slack bot commands

* Alerts on new proposals and unvoted proposals everyday at 8AM and 8PM.
   
### List of avaliable slack commands

To get response from these commands you can just use `@<bot-name> <command_name>` or `/<command_name>` from your slack workspace/channel. You will be getting response to your slack workspace/channel based on the bot token/channel ID you have configured in config.toml

    register-validator : registers the validator using chain name and validator address
    remove-validator : removes an existing validator using validator address
    list-keys : list of all the key names with network names and account addresses
    list-validators : list of all registered validators addresses with associated chains
    vote : votes on a proposal
    list-votes: lists all the votes for a given chain Id from start date to optional end date. If end date is empty, it returns the votes upto current date.
    list-commands: lists all the available commands 
    create-key : creates a new account with key name. This key name is used while voting.

## Granting authorization and funds to keys
Keys need to be funded manually and given authorization to vote in order to use them while voting.
    The granter must give the vote authorization to the grantee key before the voting can proceed.  
    The authorization to a grantee can be given by using the following command:

    For Cosmos chain:
    Usage: simd tx authz grant <grantee> <authorization_type> --msg-type <msg_type> --from <granter> [flags]
    Example: simd tx authz grant cosmos1... --msg-type /cosmos.gov.v1beta1.MsgVote --from granter

    The authorized keys can then be funded to have the ability to vote on behalf of the granter.
    The following command can be used to fund the key:
   
    simd tx bank send [from_key_or_address] [to_address] [amount] [flags]