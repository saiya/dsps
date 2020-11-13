import { Dsps, Channel as DspsChannel, Subscription as DspsSubscription } from "@dsps/client";

export const defaultChannelID = "my-first-channel";

const dsps = new Dsps({
  http: {
    baseURL: "/",
  },
});

type ChannelListListener = (channels: Channel[]) => void;

class Model {
  constructor() {
    this.newChannel("my-first-channel");
  }

  private readonly channelsMap: {
    [id: string]: Channel;
  } = {};

  private channelsListListeners: ChannelListListener[] = [];

  watchChannelsList(listener: ChannelListListener) {
    this.channelsListListeners.push(listener);
  }
  unwatchChannelsList(listener: ChannelListListener) {
    const i = this.channelsListListeners.indexOf(listener);
    if (i !== -1) this.channelsListListeners.splice(i, 1);
  }
  private channelListChanged() {
    this.channelsListListeners.forEach((h) => h(this.channels));
  }

  async newChannel(channelID: string) {
    if (this.channelsMap[channelID]) return;

    console.info("newChannel", channelID);
    const channel = new Channel(this, channelID);
    await channel._init();
    this.channelsMap[channelID] = channel;
    this.channelListChanged();
  }

  async leaveChannel(channelID: string) {
    if (!this.channelsMap[channelID]) return;

    console.info("closeChannel", channelID);
    await this.channelsMap[channelID]._close();
    delete this.channelsMap[channelID];
    this.channelListChanged();
  }

  get channels(): Channel[] {
    return Object.values(this.channelsMap).sort((a, b) => a.id.localeCompare(b.id));
  }
}

type ChatMessage = {
  at: Date;
  text: string;
};
type ChatMessagesListener = (messages: ChatMessage[]) => void;

export class Channel {
  private dspsChannel: DspsChannel;
  private dspsSubsc: DspsSubscription | null = null;

  constructor(private readonly model: Model, public readonly id: string) {
    this.dspsChannel = dsps.channel(id);
  }

  async _init() {
    this.dspsSubsc = await this.dspsChannel.subscribe({
      callback: async (msgs) => {
        this.msgs = [
          ...this.msgs,
          ...msgs.map((msg): ChatMessage => ({
            at: new Date(msg.content.at as number),
            text: msg.content.text as string,
          })),
        ];
        this.messagesChanged();
      },
      abnormalEndCallback: (e) => {
        console.error("abnormalEndCallback", e);
        // TODO: Implement UI notification
      },
    });
  }
  async _close() {
    await this.dspsSubsc?.close();
  }
  async send(message: ChatMessage) {
    this.dspsChannel.publish(null, {
      at: message.at.getTime(),
      text: message.text,
    });
  }

  private messageListeners: ChatMessagesListener[] = [];

  watchMessages(listener: ChatMessagesListener) {
    this.messageListeners.push(listener);
  }
  unwatchMessages(listener: ChatMessagesListener) {
    const i = this.messageListeners.indexOf(listener);
    if (i !== -1) this.messageListeners.splice(i, 1);
  }
  private messagesChanged() {
    this.messageListeners.forEach((h) => h(this.messages));
  }

  private msgs: ChatMessage[] = [];

  get messages(): ChatMessage[] {
    return this.msgs;
  }
}

export default new Model();
