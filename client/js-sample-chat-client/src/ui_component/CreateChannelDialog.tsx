import React from 'react';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

export default function CreateChannelDialog(props: {
  open: boolean,
  handleClose: () => void,
  handleSubscribe: (channelID: string) => void,
}) {
  const [channelID, setChannelID] = React.useState("");
  const handleClose = () => {
    setChannelID("");
    props.handleClose();
  };
  const handleEnter = () => {
    const id = channelID.trim();
    if (id === "") handleClose();
    setChannelID("");
    props.handleSubscribe(id);
  };

  return (
    <div>
      <Dialog open={props.open} onClose={handleClose} aria-labelledby="form-dialog-title">
        <DialogTitle id="form-dialog-title">Subscribe</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Enter DSPS channel ID to subscribe.
          </DialogContentText>
          <TextField
            autoFocus
            margin="dense"
            label="Channel ID"
            fullWidth
            value={channelID}
            onChange={(e) => { setChannelID(e.target.value) }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                handleEnter();
              }
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            Cancel
          </Button>
          <Button onClick={handleEnter} color="primary">
            Subscribe
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
