import os
from datetime import datetime
from io import BytesIO

import numpy as np
from fastapi import FastAPI, UploadFile
from keras.models import load_model
from PIL import Image
from skimage import transform

import script.utils

app = FastAPI()
model = load_model('model/model.h5')

def load_image_into_numpy_array(data):
    return np.array(Image.open(BytesIO(data)))

@app.get("/")
def read_root():
    return 'Hello World!'


@app.post("/image/")
async def create_upload_file(image: UploadFile):
    if not image:
        return {"message": "No upload file sent"}
    
    if not os.path.exists("upload"):
        os.mkdir("upload")

    try:
        np_image = load_image_into_numpy_array(await image.read())
        np_image = np.array(np_image).astype('float32')/255
        np_image = transform.resize(np_image, (224, 224, 3))
        np_image = np.expand_dims(np_image, axis=0)


        # image = script.utils.load_image(file_name)
        ans = model.predict(np_image)
        maping = {0 : "Neutral", 1 : "Porn", 2 : "Sexy"}
        new_ans = np.argmax(ans[0])

        return {
            "result": {
                "class": maping[new_ans],
                "percentage": round((ans[0][new_ans] * 100), 2),
            }
        }
    except:
        return {
            "result": {
                "class": 'Porn',
                "percentage": 100,
            }
        }