<script setup lang="ts">

import { Switch } from "./ui/switch"
import { Popover, PopoverTrigger, PopoverContent } from "./ui/popover"
import { Button } from "./ui/button"
import { computed, onMounted, ref, watch } from "vue"

const locations = [
    "A",
    "B",
    "C"
]
const switchStatus = ref(false)
const locationIdx = ref(0)
const locationPopoverOpen = ref(false)
const location = computed(() => locations[locationIdx.value])
const recorder = ref<MediaRecorder | undefined>()

onMounted(async () => {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    recorder.value = new MediaRecorder(stream)
})

watch(recorder, (_recorder) => {
    if (_recorder === undefined) {
        return
    }
    _recorder.ondataavailable = (e) => onRecorderDataAvailable(e.data)
})

watch(switchStatus, (on) => {
    if (on) {
        recorder.value?.start(1000)
    }
    else {
        recorder.value?.stop()
    }
})

function onRecorderDataAvailable(data: Blob) {
    data.text().then(t => console.log(t))
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
                    <Switch class="bg-zinc-500 text-zinc-500" v-model:checked="switchStatus" />
                </div>
                <div>
                    Status
                </div>
                <div>
                    {{ switchStatus ? "ON" : "OFF" }}
                </div>
            </div>
            <Popover v-model:open="locationPopoverOpen">
                <PopoverTrigger>
                    Location: {{ location }}
                </PopoverTrigger>
                <PopoverContent class="bg-white rounded-2xl w-60 border border-black">
                    <div class="grid gap-4">
                        <p class="text-sm text-muted-foreground">
                            Location for this transmitter
                        </p>
                        <div class="flex flex-col gap-2">
                            <template v-for="location, idx in locations">
                                <Button v-on:click="locationIdx = idx; locationPopoverOpen = false">{{ location
                                    }}</Button>
                            </template>
                        </div>
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    </div>
</template>
