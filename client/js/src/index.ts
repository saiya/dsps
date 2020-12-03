import "core-js/features/promise";
import "core-js/features/object";
import "core-js/features/array";
import "core-js/features/global-this";

import { DspsClient } from "./client_interface";
import { DspsClientConfig } from "./dsps_client_config";
import { DspsClientImpl } from "./internal/dsps_client";

/** `new Dsps({ ...config... })` creates {@link DspsClient} instance. */
export const Dsps: { new (config: DspsClientConfig): DspsClient } = DspsClientImpl;

export * from "./dsps_client_config";
export * from "./client_interface";
