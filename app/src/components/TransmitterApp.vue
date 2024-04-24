<script setup lang="ts">

import { Switch } from "./ui/switch"
import { Popover, PopoverTrigger, PopoverContent } from "./ui/popover"
import { Button } from "./ui/button"
import { computed, onMounted, ref, watch } from "vue"
import Pocketbase from "pocketbase"

const PB_URL = import.meta.env.PUBLIC_PB_URL
const TRANSMIT_INTERVAL_MS = parseInt(import.meta.env.PUBLIC_TRANSMIT_INTERVAL_MS)
const pb = new Pocketbase(PB_URL)

type LocationRecord = {
    id: string,
    name: string
}

type State = StateOn | StateOff

type StateOn = {
    on: true
    timerId: number
}

type StateOff = {
    on: false
}

const locations = ref<Record<string, LocationRecord>>({})
const locationId = ref<string | undefined>(undefined)
const location = computed<LocationRecord | undefined>(() => locationId.value ? locations.value[locationId.value] : undefined)
const recorder = ref<MediaRecorder | undefined>()

const transmitterState = ref<State>({on: false})
const switchOn = ref(false)
const locationPopoverOpen = ref(false)

onMounted(async () => {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    recorder.value = new MediaRecorder(stream)

    const getLocations = await pb.collection("locations").getFullList()
    locations.value = getLocations
        .map((r) => ({ id: r.id, name: r["name"] }))
        .reduce<Record<string, LocationRecord>>((obj, e) => {
            obj[e.id] = e
            return obj
        }, {})
    locationId.value = Object.values(locations.value)[0]?.id
})

watch(recorder, (_recorder) => {
    if (_recorder === undefined) {
        return
    }
    _recorder.ondataavailable = (e) => onRecorderDataAvailable(e.data)
})

watch(switchOn, (on) => {
    if (on) {
        if (!transmitterState.value.on) {
            transmitterState.value = {
                on: true,
                timerId: setInterval(() => {
                    recorder.value?.stop()
                    recorder.value?.start()
                }, TRANSMIT_INTERVAL_MS)
            }
        }
    }
    else {
        if (transmitterState.value.on) {
            clearInterval(transmitterState.value.timerId)
            transmitterState.value = { on: false }
        }
    }
})

async function onRecorderDataAvailable(data: Blob) {
    const l = location.value
    if (l === undefined) {
        return
    }
    const formData = new FormData();
    formData.append("content", data, "reading.oga")
    formData.append("location", l.id)
    try {
        await pb.collection("readings").create(formData)
    } catch (error) {
        console.error(error)
    }
}

</script>

<template>
    <div class="w-full min-h-screen bg-zinc-800 flex justify-center items-center">
        <div class="rounded-2xl p-5 bg-white w-[180px] flex flex-col gap-3 border-black border">
            <div class="text-center grid grid-cols-2 place-content-center gap-1">
                <div>
                    Enable
                </div>
                <div>
                    <Switch class="bg-zinc-500 text-zinc-500" v-model:checked="switchOn" />
                </div>
                <div>
                    Status
                </div>
                <div>
                    {{ switchOn ? "ON" : "OFF" }}
                </div>
            </div>
            <Popover v-model:open="locationPopoverOpen">
                <PopoverTrigger>
                    Location: <br /> {{ location?.name }}
                </PopoverTrigger>
                <PopoverContent class="bg-white rounded-2xl w-60 border border-black">
                    <div class="grid gap-4">
                        <p class="text-sm text-muted-foreground">
                            Location for this transmitter
                        </p>
                        <div class="flex flex-col gap-2">
                            <template v-for="location, idx in locations">
                                <Button v-on:click="locationId = idx; locationPopoverOpen = false">{{ location.name
                                    }}</Button>
                            </template>
                        </div>
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    </div>
</template>
