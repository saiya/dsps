import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import IconButton from '@material-ui/core/IconButton';
import AddIcon from '@material-ui/icons/Add';
import DeleteIcon from '@material-ui/icons/Delete';
import ChatOutlinedIcon from '@material-ui/icons/ChatOutlined';
import Snackbar from '@material-ui/core/Snackbar';
import ChannelView from "./ChannelView";
import CreateChannelDialog from "./ui_component/CreateChannelDialog";
import MuiAlert from '@material-ui/lab/Alert';
import Model, { defaultChannelID, Channel } from "./model";

const drawerWidth = 420;

const useStyles = makeStyles((theme) => ({
  root: {
    display: 'flex',
  },
  appBar: {
    width: `calc(100% - ${drawerWidth}px)`,
    marginLeft: drawerWidth,
  },
  drawer: {
    width: drawerWidth,
    flexShrink: 0,
  },
  drawerPaper: {
    width: drawerWidth,
  },
  // necessary for content to be below app bar
  toolbar: theme.mixins.toolbar,
  content: {
    flexGrow: 1,
    backgroundColor: theme.palette.background.default,
    padding: theme.spacing(3),
  },
}));

function App() {
  const [openNewChannelDialog, setNewChannelDialogOpen] = React.useState(false);
  const [currentChannelID, setCurrentChannelID] = React.useState<string | null>(defaultChannelID);
  const [channels, setChannels] = React.useState<Channel[]>(Model.channels);
  const currentChannel = channels.find((ch) => ch.id === currentChannelID);
  React.useEffect(() => {
    Model.watchChannelsList((channels) => {
      if (channels.findIndex((ch) => ch.id === currentChannelID) === -1) setCurrentChannelID(null);
      setChannels(channels);
    });
    return () => Model.unwatchChannelsList(setChannels);
  }, [setChannels, currentChannelID, setCurrentChannelID]);

  const [channelEditedMessage, setChannelEditedMessage] = React.useState<string | null>(null);

  const classes = useStyles();
  return (
    <>
      <div className={classes.root}>
        <CssBaseline />
        <AppBar position="fixed" className={classes.appBar}>
          <Toolbar>
            <IconButton color="inherit"><ChatOutlinedIcon /></IconButton>
            <Typography variant="h6" noWrap>{currentChannelID ?? "(No channel selected)"}</Typography>
          </Toolbar>
        </AppBar>
        <Drawer
          className={classes.drawer}
          variant="permanent"
          classes={{
            paper: classes.drawerPaper,
          }}
          anchor="left"
        >
          <div className={classes.toolbar} />
          <Divider />
          <List>
            {Object.values(channels).map((channel) => (
              <ListItem
                key={channel.id}
                button
                selected={channel.id === currentChannelID}
                onClick={() => setCurrentChannelID(channel.id)}
              >
                <ListItemIcon><ChatOutlinedIcon /></ListItemIcon>
                <ListItemText primary={channel.id} />
                <ListItemSecondaryAction onClick={async () => {
                  await Model.leaveChannel(channel.id);
                  setChannelEditedMessage(`Left from channel: ${channel.id}`);
                }}>
                  <IconButton edge="end"><DeleteIcon /></IconButton>
                </ListItemSecondaryAction>
              </ListItem>
            ))}
            <ListItem button onClick={() => setNewChannelDialogOpen(true)}>
              <ListItemIcon><AddIcon /></ListItemIcon>
              <ListItemText>Subscribe more channel ...</ListItemText>
            </ListItem>
          </List>
        </Drawer>
        <main className={classes.content}>
          <div className={classes.toolbar} />
          {currentChannel ? <ChannelView channel={currentChannel} /> : null}
        </main>
      </div>

      <CreateChannelDialog
        open={openNewChannelDialog}
        handleClose={() => setNewChannelDialogOpen(false)}
        handleSubscribe={async (channelID) => {
          await Model.newChannel(channelID);
          setCurrentChannelID(channelID);
          setNewChannelDialogOpen(false);
          setChannelEditedMessage(`Subscribing channel: ${channelID}`);
        }}
      />
      <Snackbar
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        autoHideDuration={3000}
        open={typeof (channelEditedMessage) === "string"}
        onClose={() => setChannelEditedMessage(null)}
      >
        <MuiAlert elevation={6} variant="filled" severity="info" onClose={() => setChannelEditedMessage(null)}>{channelEditedMessage}</MuiAlert>
      </Snackbar>
    </>
  );
}

export default App;
