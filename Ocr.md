# OCR.Z.AI

## 识别

POST：https://ocr.z.ai/api/v1/z-ocr/tasks/process

Body：form-data

| 参数名 | 参数值 | 类型 |
| ------ | ------ | ---- |
| file   | 文件   | file |

headers：

| 参数名        | 参数值                                               |
| ------------- | ---------------------------------------------------- |
| Anthorization | Bearer <Token>                                       |
| Content-Type  | multipart/form-data                                  |
| Accept        | application/json, text/plain, */*                    |
| Origin        | https://ocr.z.ai                                     |
| Referer       | https://ocr.z.ai/                                    |
| User-Agent    | Mozilla/5.0 AppleWebKit/537.36 Chrome/143 Safari/537 |

响应：

1. 成功

   ~~~json
   {
       "code": 200,
       "message": "Success",
       "data": {
           "task_id": "80ca36df-1b62-11f1-ac1b-ca6f99afd35d",
           "status": "completed",
           "file_name": "HTML+CSS.pdf",
           "file_size": 4239671,
           "file_type": "pdf",
           "file_url": "https://z-ai-audio.chatglm.cn/z_ocr/80ca36df-1b62-11f1-ac1b-ca6f99afd35d.pdf?x-oss-credential=LTAI5tCA9r3Ps8Ae4J75EWZ2%2F20260309%2Fcn-hongkong%2Foss%2Faliyun_v4_request\u0026x-oss-date=20260309T024850Z\u0026x-oss-expires=604800\u0026x-oss-signature=334728b3c406a10dece80eca08e11e6f98bc9e18f8ff715b6a40030df11da750\u0026x-oss-signature-version=OSS4-HMAC-SHA256",
           "created_at": "2026-03-09T10:48:50.72708354+08:00",
           "markdown_content": "阶段目标：掌握HTML、CSS常用布局技巧，能够独立制作网页。\n\n## 01: HTML基础",
           "json_content": "{\"md_results\":\"<markdown text>\",\"layout_details\":[{\"type\":\"text\",\"bbox\":[]}],\"data_info\":[],\"usage\":{\"pages\":1}}",
           "layout": [
               {
                   "block_content": "阶段目标：掌握HTML、CSS常用布局技巧，能够独立制作网页。",
                   "bbox": [
                       196,
                       183,
                       1113,
                       230
                   ],
                   "block_id": 0,
                   "page_index": 0,
                   "block_label": "text",
                   "score": 0
               },
               {
                   "block_content": "\u003cdiv style=\"text-align: center;\"\u003e\u003cimg src=\"https://maas-watermark-prod-new.cn-wlcb.ufileos.com/ocr%2Fcrop%2F202603091048501173f430a7c3451b%2Fcrop_1_1773024581736.png?UCloudPublicKey=TOKEN_6df395df-5d8c-4f69-90f8-a4fe46088958\u0026Signature=B1MEYVfjlXRFTGQDvrZI0jIRUnA%3D\u0026Expires=1773629381\" alt=\"Image\"/\u003e\u003c/div\u003e",
                   "bbox": [
                       217,
                       212,
                       1758,
                       637
                   ],
                   "block_id": 21,
                   "page_index": 1,
                   "block_label": "image",
                   "score": 0
               }
           ],
           "data_info": {
               "pages": [
                   {
                       "width": 1986,
                       "height": 2809
                   }
               ],
               "num_pages": 66
           }
       },
       "timestamp": 1773024582
   }
   ~~~

2. 失败

   1. 文件不符合要求

      > 文件格式不正确
      >
      > 页数超出
      >
      > 文件大小超出

      ~~~json
      {
          "code":40007,
          "message":"OCR仅支持PDF、JPG、PNG格式；文件大小限制：图片≤10MB、PDF≤50MB；PDF最大100页",
          "timestamp":1773024703
      }
      ~~~

## 对外接口

POST：/ocr/v1/files/ocr

Body：form-data

| 参数名 | 参数值 | 类型 |
| ------ | ------ | ---- |
| file   | 文件   | file |

headers：

| 参数名        | 参数值                                               |
| ------------- | ---------------------------------------------------- |
| Anthorization | Bearer <Token>                                       |
| Content-Type  | multipart/form-data                                  |
| Accept        | application/json, text/plain, */*                    |
| Origin        | https://ocr.z.ai                                     |
| Referer       | https://ocr.z.ai/                                    |
| User-Agent    | Mozilla/5.0 AppleWebKit/537.36 Chrome/143 Safari/537 |

响应：

1. 成功

   ~~~json
   {
     "task_id": "ce2641ced3e34e67b47f3b0feeb25aee",
     "message": "成功",
     "status": "succeeded",
     "words_result_num": 4,
     "words_result": [
       {
         "location": {
           "left": 79,
           "top": 122,
           "width": 1483,
           "height": 182
         },
         "words": "你好,世界!",
         "probability": {
           "average": 0.7320847511,
           "variance": 0.08768635988,
           "min": 0.3193874359
         }
       }
     ]
   }
   ~~~
   
2. 失败

   1. 文件不符合要求

      > 文件格式不正确
      >
      > 页数超出
      >
      > 文件大小超出

      ~~~json
      {
          "code":40007,
          "message":"OCR仅支持PDF、JPG、PNG格式；文件大小限制：图片≤10MB、PDF≤50MB；PDF最大100页",
          "timestamp":1773024703
      }
      ~~~

转换对照表：

| 对外字段              | OCR.Z.AI字段  | 转换规则     |
| --------------------- | ------------- | ------------ |
| words                 | block_content | 直接赋值     |
| location.left         | bbox[0]       | x1           |
| location.top          | bbox[1]       | y1           |
| location.left + width | bbox[2]       | x2           |
| location.top + height | bbox[3]       | y2           |
| probability.average   | score         | 置信度       |
| index                 | block_id      | 顺序编号     |
| page                  | page_index    | 单页OCR默认0 |
| 固定值                | block_label   | text         |

# 四、bbox转换公式

GLM-5写代码时最容易错这里。

智谱：

```
left, top, width, height
```

你接口：

```
x1, y1, x2, y2
```

转换：

```
x1 = left
y1 = top
x2 = left + width
y2 = top + height
```

------

# 五、layout生成规则

伪代码给 GLM-5：

```
layout = []

for i, item in enumerate(words_result):
    left = item["location"]["left"]
    top = item["location"]["top"]
    width = item["location"]["width"]
    height = item["location"]["height"]

    bbox = [
        left,
        top,
        left + width,
        top + height
    ]

    layout.append({
        "block_content": item["words"],
        "bbox": bbox,
        "block_id": i,
        "page_index": 0,
        "block_label": "text",
        "score": item["probability"]["average"]
    })
```

------

# 六、markdown_content生成

最简单方式：

```
markdown_content = "\n".join(
    item["words"] for item in words_result
)
```

如果想稍微高级一点：

```
markdown_content = "\n\n".join(
    item["words"] for item in words_result
)
```

------

# 七、json_content生成

```
json_content = {
    "md_results": markdown_content,
    "layout_details": layout,
    "data_info": {},
    "usage": {
        "pages": 1
    }
}
```

------

# 八、data_info生成

```
data_info = {
    "pages": [
        {
            "width": image_width,
            "height": image_height
        }
    ],
    "num_pages": 1
}
```

------

# 九、最终结构

最终 OCR.Z.AI：

```
OCRResult
 ├ markdown_content
 ├ json_content
 ├ layout
 └ data_info
```

------

# 十、给 GLM-5 的最重要规则

```
1. 智谱 OCR words_result → OCR.Z.AI layout
2. location(left,top,width,height) → bbox(x1,y1,x2,y2)
3. words → block_content
4. probability.average → score
5. markdown_content = words_result 拼接
```

基本就不会写错。
