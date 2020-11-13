import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import InputBase from '@material-ui/core/InputBase';
import IconButton from '@material-ui/core/IconButton';
import ChatOutlinedIcon from '@material-ui/icons/ChatOutlined';
import SendIcon from '@material-ui/icons/Send';
import { Channel } from "./model";

const useStyles = makeStyles((theme) => ({
  root: {},
  message: {
    marginBottom: "1em",
  },
  messageIcon: {
    marginLeft: "5pt",
  },
  timestamp: {
    color: "gray",
    marginRight: "7pt",
  },
  inputForm: {
    display: 'flex',
  },
  input: {
    flex: 1,
    marginLeft: "7pt",
  },
  sendButton: {
    padding: 10,
  },
}));

export default function ChannelView(props: {
  channel: Channel;
}) {
  const { channel } = props;

  const [messages, setMessages] = React.useState(channel.messages);
  React.useEffect(() => {
    channel.watchMessages(setMessages);
    return () => channel.unwatchMessages(setMessages);
  }, [channel, setMessages]);

  const [msgText, setMsgText] = React.useState("");
  const send = async () => {
    if (msgText.trim() === "") return;
    channel.send({ at: new Date(), text: msgText });
  };

  const classes = useStyles();
  return <div className={classes.root}>
    {messages.map((msg, i) => <div className={classes.message}>
      <Paper key={i}>
        <Grid container direction="row" alignItems="center" spacing={1}>
          <Grid item className={classes.messageIcon}>
            <ChatOutlinedIcon color="disabled" />
          </Grid>
          <Grid item xs>
            {msg.text}
          </Grid>
          <Grid item className={classes.timestamp}>
            {`${msg.at}`}
          </Grid>
        </Grid>
      </Paper>
    </div>)}
    <Paper component="form" className={classes.inputForm}>
      <InputBase
        autoFocus={true}
        className={classes.input}
        placeholder="Message..."
        value={msgText}
        onChange={(e) => { setMsgText(e.target.value) }}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            send();
            e.preventDefault();
          }
        }}
      />
      <IconButton type="submit" className={classes.sendButton} onClick={send}>
        <SendIcon color="primary" />
      </IconButton>
    </Paper>
  </div>;
}
