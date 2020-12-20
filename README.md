# Earth

## What does this do?

Generate a video of the Earth spinning from images taken from the [DSCOVR: EPIC
](https://epic.gsfc.nasa.gov/) spacecraft!

## Overview

The Deep Space Climate Observatory (DSCOVR) is a spacecraft launched by SpaceX in February of 2015. There's a camera onboard called EPIC (Earth Polychromatic Imaging Camera) that takes photos of Earth. Here's one:

<img alt="earth_example_image" src="https://user-images.githubusercontent.com/21317789/102694332-52140e00-41ee-11eb-8fd9-0783b91481a3.png" width="512">

Yes, that's a real photo! Amazing, right? This spacecraft is a **MILLION** miles from Earth. 🤯
It sits at the [Lagrange point](https://en.wikipedia.org/wiki/Lagrange_point) (specifically, L1), which means it's in a pocket in space where the gravitational forces from the Sun and the Earth are about the same in opposite directions, canceling each other out. In short, it's in a semi-stable position in space very far from Earth.

![Lagrangianpointsanimated](https://user-images.githubusercontent.com/21317789/102694571-f8acde80-41ef-11eb-8971-5b98a45e648d.gif)

Photos are taken roughly every two hours and are publicly available on the web. I don't know about you, but I find the photograph above to be incredibly beautiful. NASA's website provides some [galleries](https://epic.gsfc.nasa.gov/galleries) where some truly amazing things have been captured over the years.
Here's a lunar transit where the moon passes between the Earth and the camera. That's the far side of the moon you're seeing! Something that is never seen from Earth because of the moon's [tidally locked orbit](https://en.wikipedia.org/wiki/Tidal_locking).

<img width="1286" alt="image" src="https://user-images.githubusercontent.com/21317789/102694700-c3ed5700-41f0-11eb-94d7-3760169f56a3.png">

Here's an image of the California wildfires from September 2020.

![wildfires](https://user-images.githubusercontent.com/21317789/102694807-686f9900-41f1-11eb-9c12-a6b7cf99a880.png)

---

After some time browsing these images, I decided it'd be neat if I could download a bunch of photos and create a video of the Earth spinning!

## First attempt

First I threw together a quick script to download some photos using the [API](https://epic.gsfc.nasa.gov/about/api). I wanted to start small and just pick a one or two day span of time -- I picked sometime in September in 2018 (9/6 - 9/8).
There's a nifty CLI tool called [`ffmpeg`](https://ffmpeg.org/) that I knew I could use to convert a bunch of PNGs into an MP4. I ran `ffmpeg` and fed it my downloaded photos but the results were disappointing.
`ffmpeg` allows you to define a frame-rate for the video output. Here's the result at 2 frames per second (Github doesn't allow for video within a README, and converting them to GIFs produces files too large to upload, so you'll have to click through on the video links).

[2 FPS video](https://lindley-earth.s3.amazonaws.com/earth_2fps.mp4)

Not very nice, huh? It's awfully choppy because the frame-rate is low. So let's try bumping it up to 60 frames per second:

[60 FPS video](https://lindley-earth.s3.amazonaws.com/earth_60fps.mp4)

While this makes the video "smoother" because there are more frames per second, it makes the video shorter and also way more chaotic. I was looking for a more relaxing watching experience where you could take in the sights.
The reality here is that there just aren't enough "frames", or photos for me to create the video I was hoping for. And we can't just create new photos from thin air.

Or can we? 🤔

## Interpolation

In case you're not familiar with the concept of interpolation, it involves estimating some new information from existing information.
For example, if it's 40℉ at 11 am, and 60℉ at 1 pm, then you can _interpolate_ that it was probably 50℉ at noon. Make sense?
You can interpolate all sorts of information -- even visual information!

Digital images on a screen can be thought of as a lattice of [pixels](https://en.wikipedia.org/wiki/Pixel). Displays these days have a **lot** of pixels that may even be too small to see with the naked eye. If you're old enough you will remember displays where you _could_ actually pick out individual pixels.

<img alt="mario" src="https://user-images.githubusercontent.com/21317789/102695636-9c00f200-41f6-11eb-9253-314edfcbd20d.png" width="40">

The pixels are laid out in a grid on your screen, with each pixel assigned some color value. Let's take a simple example of some screen displaying an image with one black pixel.

<img width="296" alt="image" src="https://user-images.githubusercontent.com/21317789/102695868-18e09b80-41f8-11eb-83a3-e28aea455e2e.png">

Now let's imagine we had a second image where the pixel is 4 cells further to the right. So, you could say it has "moved" to the right.

<img width="296" alt="image" src="https://user-images.githubusercontent.com/21317789/102695884-33b31000-41f8-11eb-967b-c666d9193e6d.png">

If we imagine this pixel having some constant "speed", we could interpolate that it was probably at the midpoint (2 cells to the right) halfway through its journey.

<img width="296" alt="image" src="https://user-images.githubusercontent.com/21317789/102695919-7674e800-41f8-11eb-831a-0b082f77b7da.png">

We can also imagine doing the same thing in two dimensions. Below I've just merged the three points into a single image to illustrate the concept more succinctly. Here, `0` is the start, `1` is the end, and we've computed some midpoint.

<img width="296" alt="image" src="https://user-images.githubusercontent.com/21317789/102695966-dcfa0600-41f8-11eb-82db-88192cb99405.png">

As I mentioned before, pixels have a color value. For many displays, a pixel is actually made up of three separate tiny lights, or diodes. For LED displays, these diodes are typically red, green, and blue in color.

<img width="300" alt="diodes" src="https://user-images.githubusercontent.com/21317789/102696082-b12b5000-41f9-11eb-96b6-c7da56cd5b56.jpg">

The intensity of each diode in each pixel can be controlled individually. From a distance, your brain can't pick them apart, and they are mushed together and *perceived* as a single color. For instance, if the red light is at 80% and the blue light is at 50% and the green light is at 0% this will produce some dark reddish-purple.

How your brain might percieve that pixel:

<img width="123" alt="image" src="https://user-images.githubusercontent.com/21317789/102696324-8215de00-41fb-11eb-8579-4c1841933952.png">

Color can also be interpolated, as the intensities of a pixel's red, green, and blue values shift from one state to another.

<img width="587" alt="image" src="https://user-images.githubusercontent.com/21317789/102696296-4e3ab880-41fb-11eb-9ee8-c3f9e0fd5812.png">

Put that together with motion and that might look something like this:

<img width="298" alt="image" src="https://user-images.githubusercontent.com/21317789/102696361-bc7f7b00-41fb-11eb-8c11-b09847082589.png">

There's no reason why these same concepts could be applied to every single pixel of a much larger image.

By now you might see where this is going... The idea is that we could theoretically take two images of the Earth, and interpolate a third image in between the two to create a new frame that didn't exist before.

## Spheres

We know the Earth to be a sphere. You're not in that other camp, are you?
And the spherical shape of the earth introduces a particular challenge to our interpolation efforts. You see, we can think of each point on earth generally moving in a straight path as it spins. However, straight paths on the surface of a sphere often don't appear straight at all. What do I mean by this?

Imagine you and a friend are standing at the equator looking north but hundreds of miles away from each other. You both walk in a "straight" line towards the north pole. If you started on parallel paths, then how is it that your paths intersect at the north pole?
Well, it's because you're walking along a curved plane.

<img width="300" alt="geo-desic-poles" src="https://user-images.githubusercontent.com/21317789/102696888-ad023100-41ff-11eb-9599-41d5b9800aef.gif">

Even though you've taken a straight path, it appears to be curved when projected onto a two-dimensional plane. By the way, straight paths along curved surfaces are called [geodesics](https://en.wikipedia.org/wiki/Geodesic) and it's also why a lot of flight paths *appear* to be curved when viewed on a map. Yes, your airplane is taking the most direct path, the curve is only an illusion because the globe is being flattened onto a 2D image.

![flight](https://user-images.githubusercontent.com/21317789/102696949-29950f80-4200-11eb-9efd-50dd44e1ac96.jpg)

---

What all this means for our interpolation is that we need to employ some good old trigonometry to work out the path of a particular pixel.

<img width="300" alt="geo-desic" src="https://user-images.githubusercontent.com/21317789/102696716-36186880-41fe-11eb-9b4d-311b491e88d8.png">

Another consideration is that the spacecraft isn't exactly still. That gravitational balance is delicate and it's making adjustments constantly. This means that sometimes it's closer to Earth than other times, making the Earth appear *larger* in the frame when it's closer. So not only do we need to apply a transform to these images but also scale them so that the transition between the frames is seamless and believable. Thankfully, NASA provides metadata about each image that includes the x,y,z coordinates of the spacecraft which should allow us to compute how far away from the Earth it is for each photo. And, you guessed it, we can interpolate those missing distance values as well!

Here's a GIF of the spacecraft's trajectory taken from the [Deep Space Climate Observatory Wikipedia page](https://en.wikipedia.org/wiki/Deep_Space_Climate_Observatory)

![Animation_of_Deep_Space_Climate_Observatory_trajectory](https://user-images.githubusercontent.com/21317789/102702007-8659ef80-422b-11eb-95f3-34dea7497749.gif)

## Result

What if I told you this image of Earth was never actually taken?

<img width="512" alt="fake earth" src="https://user-images.githubusercontent.com/21317789/102697252-81cd1100-4202-11eb-8dda-f1c399e1e61b.png">

Yep, it's fake! It's the computed midpoint between the two **real** images on the far left and right.

<img width="1798" alt="image" src="https://user-images.githubusercontent.com/21317789/102697300-d8d2e600-4202-11eb-8a46-43a0761c496d.png">

Looks pretty believable, I think. Once I got the interpolation worked out, I went ahead and generated a bunch of fake frames between each photo over a two-day span. Finally, with all these frames I was able to stitch them together into the video I've been after!

### [Final video](https://lindley-earth.s3.amazonaws.com/earth.mp4)

In this video, only 26 of the 1,396 frames are real photos. Or, put another way, it's 98% fake!

## Code

### Dependencies

- [Go](https://golang.org/)
- [ffmpeg](https://ffmpeg.org/)

### Quick start

Run `go install` within the directory.

Then run `earth`.

**Be aware:** the program takes several minutes to complete using the default start and end date (which is only a two-day interval). Although the program downloads a fair number of images, there are far more generated images which means this process is CPU bound. You'll see what I mean when you hear your fan start to spin up. All these generated images will take up a non-trivial amount of disk space, so keep an eye on that if you choose to broaden the date range.

The general strategy for the program is to build a pipeline that executes the following steps concurrently:

- Download the image metadata from NASA starting at some provided date.
- Pass each image metadata off to a worker that downloads the actual images.
- The downloaded images are processed in pairs with N frames are generated between each one.
- All the images (real and fake) are piped to `ffmpeg` to generate a video.

I tried to make the pipeline robust/efficient, in that it can be canceled and properly drained. And each step will skip over any downloads or generated frames that have already been saved to disk. Meaning you can stop the program at any point and then run it again to pick back up where you left off.

This concurrent pipeline pattern is one that I had read about it in [*Concurrency in Go*](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/) by Katherine Cox-Buday.

Shout out to [`illectro`](https://github.com/illectro): your python implementation helped me wrap my head around the trickier bits of the interpolation steps.

## Blooper reel

![blooper images](https://user-images.githubusercontent.com/21317789/102702891-df2e8580-4235-11eb-8be0-1797ba5a8aab.png)