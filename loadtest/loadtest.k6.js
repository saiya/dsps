import http from 'k6/http';
import { group, sleep, check } from 'k6';
import { Counter, Trend } from 'k6/metrics';

//
// To run in local, use ./experiment-local.sh
// See experiment logbook (e.g. ./experiments/20201229-on-macbook/README.md) for steps to reproduce.
//


//
// --- Compute client role & parameter ---
//

const settings = {
    BASE_URL: __ENV.BASE_URL || "http://localhost:3000",
    CHANNEL_ID_PREFIX: __ENV.CHANNEL_ID_PREFIX || "",

    TEST_RAMPUP_SEC: __ENV.TEST_RAMPUP_SEC || 1,
    TEST_DURATION_SEC: __ENV.TEST_DURATION_SEC || 3,
    TEST_RAMPDOWN_SEC: __ENV.TEST_RAMPUP_SEC || 0,

    CHANNELS: +(__ENV.CHANNELS) || 50,
    PUBLISHER_PER_CHANEL: +(__ENV.PUBLISHER_PER_CHANEL) || 1,
    SUBSCRIBER_PER_PUBLISHER: +(__ENV.SUBSCRIBER_PER_PUBLISHER) || 3,

    PUBLISH_MESSAGES_PER_ITERATION: +(__ENV.PUBLISH_MESSAGES_PER_ITERATION) || 3,

    PUBLISH_INTERVAL_SEC: +(__ENV.PUBLISH_INTERVAL_SEC) || 1,
    SUBSCRIBE_ACK_WAIT_SEC: +(__ENV.SUBSCRIBE_INERVAL_SEC) || 0.01,
    SUBSCRIBE_INERVAL_SEC: +(__ENV.SUBSCRIBE_INERVAL_SEC) || 0.01,
}
const clientsPerChannel = (1 + settings.SUBSCRIBER_PER_PUBLISHER) * settings.PUBLISHER_PER_CHANEL;
const targetVU = settings.CHANNELS * clientsPerChannel;
if (__VU === 1) console.log(`clientsPerChannel: ${clientsPerChannel}, targetVU: ${targetVU}, settings: ${JSON.stringify(settings, null, 2)}`);

const channelNumber = Math.floor((__VU - 1) / clientsPerChannel);
const isPublisher = (((__VU - 1) % clientsPerChannel) < settings.PUBLISHER_PER_CHANEL);

//
// --- Custom metrics ---
//
const fetchedMessagesCounter = new Counter('dsps_fetched_messages');
const ttfbTrends = { // TTFB = time to first byte
    publish: new Trend('dsps_ttfb_ms_publish'),
    ack: new Trend('dsps_ttfb_ms_ack'),
};
const messageDelayTrends = new Trend('dsps_msg_delay_ms');

//
// --- k6 config ---
//
export const options = {
    stages: [
        { duration: `${settings.TEST_RAMPUP_SEC}s`, target: targetVU },
        { duration: `${settings.TEST_DURATION_SEC}s`, target: targetVU },
        { duration: `${settings.TEST_RAMPDOWN_SEC}s`, target: 0 },
    ],
    thresholds: {  // https://k6.io/docs/using-k6/thresholds
        checks: ['rate >= 0.9999'],
        dsps_fetched_messages: [`count >= ${0.9 * (targetVU * settings.TEST_DURATION_SEC) / settings.PUBLISH_INTERVAL_SEC}`],
    },
};

//
// --- Load test implementation ---
//
export function setup() {
    const chars = "abcdefghijklmnopqrstuvwxyz".split("");
    let randomID = "";
    for (let i = 0; i < 16; i++) {
        randomID += chars[Math.floor(chars.length * Math.random())];
    }

    const data = {
        randomID: randomID,
    };
    console.log(`data: ${JSON.stringify(data)}`);
    return data;
}

export default function (data) {
    const { randomID } = data;
    const channelID = `${settings.CHANNEL_ID_PREFIX}${randomID}-${channelNumber}`;
    const subscriberID = `sbsc-${__VU}`;

    if (isPublisher) {
        group("publisher", () => {
            publisherScenario(channelID);
        });
    } else {
        group("subscriber", () => {
            subscriberScenario(channelID, subscriberID);
        });
    }
}

function publisherScenario(channelID) {
    for (let i = 0; i < settings.PUBLISH_MESSAGES_PER_ITERATION; i++) {
        publishMessage(channelID, `${__VU}-${__ITER}-${i}`);
    }
    sleep(settings.PUBLISH_INTERVAL_SEC);
}

function subscriberScenario(channelID, subscriberID) {
    if (__ITER === 0) {
        createSubscription(channelID, subscriberID);
    }

    let ackHandle;
    group("fetch", () => {
        ackHandle = fetchMessages(channelID, subscriberID);
    });
    group("ack", () => {
        if (ackHandle) {
            sleep(settings.SUBSCRIBE_ACK_WAIT_SEC);
            consumeAckHandle(channelID, subscriberID, ackHandle);
        }
    });

    sleep(settings.SUBSCRIBE_INERVAL_SEC);
}

function publishMessage(channelID, messageID) {
    const message = createMessage();
    const res = http.put(`${settings.BASE_URL}/channel/${channelID}/message/${messageID}`, message, { tags: { endpoint: "publish" } });
    check(res, {
        "is status 200": (r) => r.status === 200,
    });
    ttfbTrends.publish.add(res.timings.waiting);
}

function createSubscription(channelID, subscriberID) {
    const res = http.put(`${settings.BASE_URL}/channel/${channelID}/subscription/polling/${subscriberID}`, undefined, { tags: { endpoint: "createSubscription" } });
    check(res, {
        "is status 200": (r) => r.status === 200,
    });
}

function fetchMessages(channelID, subscriberID) {
    const res = http.get(`${settings.BASE_URL}/channel/${channelID}/subscription/polling/${subscriberID}?timeout=3s`, undefined, { tags: { endpoint: "fetch" } });
    const receivedAt = new Date();

    let body = null;
    let isValidJSON = true;
    try {
	body = res.json();
    }catch(e){
	console.log(`Invalid response ${res.status} (${JSON.stringify(res.headers)}): ${res.body}`);
	isValidJSON = false;
    }
    if(!check(res, {
	"returns valid JSON": () => isValidJSON,
        "is status 200": (r) => r.status === 200,
        "has messages array": (r) => (typeof(body) === "object" && typeof (body.messages) === "object" && typeof (body.messages.length) === "number"),
    })) return null;
    if (body.messages) {
        fetchedMessagesCounter.add(body.messages.length);
        for (let i = 0; i < body.messages.length; i++) {
            readMessage(receivedAt, body.messages[i]);
        }
    }
    return body.ackHandle;
}

function consumeAckHandle(channelID, subscriberID, ackHandle) {
    const res = http.del(`${settings.BASE_URL}/channel/${channelID}/subscription/polling/${subscriberID}/message?ackHandle=${ackHandle}`, undefined, { tags: { endpoint: "ack" } });
    check(res, {
        "is status 204": (r) => r.status === 204,
    });
    ttfbTrends.ack.add(res.timings.waiting);
}

function createMessage() {
    return JSON.stringify({
        at: (new Date()).getTime(),
    });
}

function readMessage(receivedAt, msg) {
    messageDelayTrends.add(receivedAt.getTime() - msg.content.at);
}
