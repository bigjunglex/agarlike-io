#!/usr/bin/env -S deno run --allow-all

/**
 * Bot runner for smoke tests + flooding with simulated clients
 * util executable
*/

/**
 * hardcorded clients to register / login as 
 */
const CLIENT_CREDS: Record<string, string> = {
    "BotBomber"     : "123456qwe",
    "BotBoss"       : "123456rwwe",
    "BotBoat"       : "123456qwe",
    "BotBoast"      : "123456eqw",
    "BotBridge"     : "123456eqw",
    "BotBrother"    : "123456eqq",
    "BotBread"      : "123456eqq",
} 
const SERVER_URL = "http://localhost:8075/ws"


class BotRunner {
    private clients: Record<number, WebSocket>  =  {}
    private size: number

    constructor(size: number) {
        this.size = size
    }


}

